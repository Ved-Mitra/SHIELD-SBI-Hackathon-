package org.example.shield.gate

expect class AppAttestation() {
    suspend fun verifyAppIntegrity(): AttestationResult
}

sealed class AttestationResult {
    object Success: AttestationResult()
    data class Failure(val error: String): AttestationResult()
}

