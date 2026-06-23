package org.example.shield.gate

expect class AppAttestation() {
    suspend fun verifyAppIntegrity(nonce: String): AttestationResult
}

sealed class AttestationResult {
    data class Success(val token: String): AttestationResult()
    data class Failure(val error: String): AttestationResult()
}

