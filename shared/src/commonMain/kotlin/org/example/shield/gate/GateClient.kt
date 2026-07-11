package org.example.shield.gate

expect class GateClient() {
    suspend fun getGate1Token(integrityToken: String, nonce: String): Result<String>
    suspend fun getGate2Token(gate1Token: String): Result<String>
    suspend fun authenticateGate3(gate2Token: String, username: String): Result<String>
    suspend fun registerGate3(gate2Token: String, username: String): Result<String>
}
