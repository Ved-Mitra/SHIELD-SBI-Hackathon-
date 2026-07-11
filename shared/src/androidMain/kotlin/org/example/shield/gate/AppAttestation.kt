package org.example.shield.gate

import kotlinx.coroutines.delay
import kotlinx.coroutines.tasks.await
import org.example.shield.AndroidContextProvider
import com.google.android.play.core.integrity.IntegrityManagerFactory
import com.google.android.play.core.integrity.IntegrityTokenRequest

actual class AppAttestation actual constructor() {
    actual suspend fun verifyAppIntegrity(nonce: String): AttestationResult {
        return try {
            val integrityManager = IntegrityManagerFactory.create(AndroidContextProvider.context)
            val request = IntegrityTokenRequest.builder()
                .setCloudProjectNumber(428692187318)
                .setNonce(nonce)
                .build()
            
            val tokenResponse = integrityManager.requestIntegrityToken(request).await()
            AttestationResult.Success(tokenResponse.token())
        } catch (e: Exception) {
            AttestationResult.Failure(e.message ?: "Play Integrity API failed")
        }
    }
}
