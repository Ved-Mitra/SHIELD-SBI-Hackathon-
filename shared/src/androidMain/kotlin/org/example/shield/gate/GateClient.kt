package org.example.shield.gate

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.json.JSONObject
import java.io.OutputStreamWriter
import java.net.HttpURLConnection
import java.net.URL

actual class GateClient actual constructor() {

    // Using 10.0.2.2 to access localhost from Android Emulator
    private val GATE1_URL = "http://10.0.2.2:8081/gate1/attest"
    private val GATE2_URL = "http://10.0.2.2:8080/gate2/token"
    private val GATE3_BEGIN_URL = "http://10.0.2.2:8082/gate3/authenticate/begin"
    private val GATE3_FINISH_URL = "http://10.0.2.2:8082/gate3/authenticate/finish"

    actual suspend fun getGate1Token(integrityToken: String, nonce: String): Result<String> = withContext(Dispatchers.IO) {
        try {
            val url = URL(GATE1_URL)
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
            // Fallback for prototype testing without backend running
            Result.success("mock_g1_jwt_token")
        }
    }

    actual suspend fun getGate2Token(gate1Token: String): Result<String> = withContext(Dispatchers.IO) {
        try {
            val url = URL(GATE2_URL)
            val conn = url.openConnection() as HttpURLConnection
            conn.requestMethod = "POST"
            // mTLS requires an SSLSocketFactory injected here, bypassed for prototype HTTP fallback
            conn.setRequestProperty("Authorization", "Bearer $gate1Token")
            
            if (conn.responseCode == 200) {
                val response = conn.inputStream.bufferedReader().use { it.readText() }
                val json = JSONObject(response)
                Result.success(json.getString("token"))
            } else {
                Result.failure(Exception("Gate 2 failed with code ${conn.responseCode}"))
            }
        } catch (e: Exception) {
            // Fallback for prototype testing without backend running
            Result.success("mock_g2_jwt_token")
        }
    }

    actual suspend fun authenticateGate3(gate2Token: String, username: String): Result<String> = withContext(Dispatchers.IO) {
        try {
            // Step 1: Begin
            val beginUrl = URL(GATE3_BEGIN_URL)
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

            // In a real app, we pass the beginConn response to Android CredentialManager API here
            // to prompt FaceID/Fingerprint and generate the signed assertion.

            // Simulate hardware biometric prompt delay
            kotlinx.coroutines.delay(1500)

            // Step 2: Finish
            val finishUrl = URL("$GATE3_FINISH_URL?username=$username")
            val finishConn = finishUrl.openConnection() as HttpURLConnection
            finishConn.requestMethod = "POST"
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
            // Fallback for prototype testing
            Result.success("mock_final_banking_session_token")
        }
    }
}
