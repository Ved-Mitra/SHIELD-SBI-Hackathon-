package org.example.shield.gate

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.example.shield.AndroidContextProvider
import org.json.JSONObject
import java.io.InputStream
import java.io.OutputStreamWriter
import java.net.HttpURLConnection
import java.net.URL
import java.security.KeyStore
import javax.net.ssl.HttpsURLConnection
import javax.net.ssl.KeyManagerFactory
import javax.net.ssl.SSLContext
import javax.net.ssl.TrustManagerFactory

actual class GateClient actual constructor() {

    actual suspend fun getGate1Token(integrityToken: String, nonce: String): Result<String> = withContext(Dispatchers.IO) {
        try {
            val url = URL(Constants.GATE1_URL)
            val conn = url.openConnection() as HttpURLConnection
            conn.requestMethod = "POST"
            conn.setRequestProperty("Content-Type", "application/json")
            conn.doOutput = true

            val payload = JSONObject().apply {
                put("platform", "android")
                put("integrity_token", integrityToken)
                put("nonce", nonce)
            }

            OutputStreamWriter(conn.outputStream).use { it.write(payload.toString()) }

            if (conn.responseCode == 200) {
                val response = conn.inputStream.bufferedReader().use { it.readText() }
                val json = JSONObject(response)
                Result.success(json.getString("token"))
            } else {
                Result.failure(Exception("Gate 1 failed with code ${conn.responseCode}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    actual suspend fun getGate2Token(gate1Token: String): Result<String> = withContext(Dispatchers.IO) {
        try {
            val url = URL(Constants.GATE2_URL)
            val conn = url.openConnection() as HttpsURLConnection
            conn.requestMethod = "POST"
            
            // ── Real mTLS Configuration ──
            // Load the PKCS12 client certificate from Android res/raw
            // The developer must place their client.p12 file in androidApp/src/main/res/raw/client.p12
            val context = AndroidContextProvider.context
            val resId = context.resources.getIdentifier("client", "raw", context.packageName)
            if (resId != 0) {
                val keyStore = KeyStore.getInstance("PKCS12")
                context.resources.openRawResource(resId).use { inputStream ->
                    keyStore.load(inputStream, "password".toCharArray()) // Replace "password" with real p12 password
                }
                
                val keyManagerFactory = KeyManagerFactory.getInstance(KeyManagerFactory.getDefaultAlgorithm())
                keyManagerFactory.init(keyStore, "password".toCharArray())
                
                // For hackathon local dev, we trust all server certs because Envoy uses self-signed certs
                val trustAllCerts = arrayOf<javax.net.ssl.TrustManager>(object : javax.net.ssl.X509TrustManager {
                    override fun getAcceptedIssuers(): Array<java.security.cert.X509Certificate>? = arrayOf()
                    override fun checkClientTrusted(certs: Array<java.security.cert.X509Certificate>, authType: String) {}
                    override fun checkServerTrusted(certs: Array<java.security.cert.X509Certificate>, authType: String) {}
                })
                
                val sslContext = SSLContext.getInstance("TLS")
                sslContext.init(keyManagerFactory.keyManagers, trustAllCerts, java.security.SecureRandom())
                conn.sslSocketFactory = sslContext.socketFactory
                conn.hostnameVerifier = javax.net.ssl.HostnameVerifier { _, _ -> true }
            } else {
                throw Exception("mTLS client.p12 certificate not found in res/raw")
            }
            
            conn.setRequestProperty("Authorization", "Bearer $gate1Token")
            
            if (conn.responseCode == 200) {
                val response = conn.inputStream.bufferedReader().use { it.readText() }
                val json = JSONObject(response)
                Result.success(json.getString("token"))
            } else {
                Result.failure(Exception("Gate 2 failed with code ${conn.responseCode}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    actual suspend fun authenticateGate3(gate2Token: String, username: String): Result<String> = withContext(Dispatchers.IO) {
        try {
            // Step 1: Begin
            val beginUrl = URL(Constants.GATE3_BEGIN_URL)
            val beginConn = beginUrl.openConnection() as HttpURLConnection
            beginConn.requestMethod = "POST"
            beginConn.setRequestProperty("Authorization", "Bearer $gate2Token")
            beginConn.setRequestProperty("Content-Type", "application/json")
            beginConn.doOutput = true

            val beginPayload = JSONObject().apply { put("username", username) }
            OutputStreamWriter(beginConn.outputStream).use { it.write(beginPayload.toString()) }

            if (beginConn.responseCode != 200) {
                return@withContext Result.failure(Exception("Gate 3 Begin failed: ${beginConn.responseCode}"))
            }

            // In a real FIDO2 app, we pass the beginConn response to Android CredentialManager API here.
            // For the hackathon demonstration, we trigger the system biometric/PIN prompt
            // to show the user interaction, then send a mock signed assertion.
            val biometricSuccess = BiometricHelper.promptBiometric()
            if (!biometricSuccess) {
                return@withContext Result.failure(Exception("Biometric Authentication Cancelled or Failed"))
            }

            // Step 2: Finish
            val finishUrl = URL("${Constants.GATE3_FINISH_URL}?username=$username")
            val finishConn = finishUrl.openConnection() as HttpURLConnection
            finishConn.requestMethod = "POST"
            finishConn.setRequestProperty("Authorization", "Bearer $gate2Token")
            finishConn.setRequestProperty("Content-Type", "application/json")
            finishConn.doOutput = true

            val finishPayload = JSONObject().apply {
                // Mock signed assertion
                put("id", "credential_id")
                put("type", "public-key")
            }
            OutputStreamWriter(finishConn.outputStream).use { it.write(finishPayload.toString()) }

            if (finishConn.responseCode == 200) {
                val response = finishConn.inputStream.bufferedReader().use { it.readText() }
                val json = JSONObject(response)
                Result.success(json.getString("session_token"))
            } else {
                Result.failure(Exception("Gate 3 Finish failed: ${finishConn.responseCode}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    actual suspend fun registerGate3(gate2Token: String, username: String): Result<String> = withContext(Dispatchers.IO) {
        try {
            // Step 1: Begin Registration
            val beginUrl = URL(Constants.GATE3_REGISTER_BEGIN_URL)
            val beginConn = beginUrl.openConnection() as HttpURLConnection
            beginConn.requestMethod = "POST"
            beginConn.setRequestProperty("Authorization", "Bearer $gate2Token")
            beginConn.setRequestProperty("Content-Type", "application/json")
            beginConn.doOutput = true

            val beginPayload = JSONObject().apply { put("username", username) }
            OutputStreamWriter(beginConn.outputStream).use { it.write(beginPayload.toString()) }

            if (beginConn.responseCode != 200) {
                return@withContext Result.failure(Exception("Gate 3 Registration Begin failed: ${beginConn.responseCode}"))
            }

            // Prompt user for fingerprint/face to generate a new Passkey
            val biometricSuccess = BiometricHelper.promptBiometric()
            if (!biometricSuccess) {
                return@withContext Result.failure(Exception("Biometric Registration Cancelled or Failed"))
            }

            // Step 2: Finish Registration
            val finishUrl = URL("${Constants.GATE3_REGISTER_FINISH_URL}?username=$username")
            val finishConn = finishUrl.openConnection() as HttpURLConnection
            finishConn.requestMethod = "POST"
            finishConn.setRequestProperty("Authorization", "Bearer $gate2Token")
            finishConn.setRequestProperty("Content-Type", "application/json")
            finishConn.doOutput = true

            val finishPayload = JSONObject().apply {
                // Mock FIDO2 public key creation response
                put("id", "new_credential_id")
                put("type", "public-key")
            }
            OutputStreamWriter(finishConn.outputStream).use { it.write(finishPayload.toString()) }

            if (finishConn.responseCode == 200 || finishConn.responseCode == 201) {
                Result.success("registration_success")
            } else {
                Result.failure(Exception("Gate 3 Registration Finish failed: ${finishConn.responseCode}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
