package org.example.shield.gate

import kotlinx.coroutines.delay

actual class GateClient actual constructor() {
    actual suspend fun getGate1Token(integrityToken: String, nonce: String): Result<String> {
        delay(500)
        return Result.success("mock_g1_jwt_token_ios")
    }

    actual suspend fun getGate2Token(gate1Token: String): Result<String> {
        delay(500)
        return Result.success("mock_g2_jwt_token_ios")
    }

    actual suspend fun authenticateGate3(gate2Token: String, username: String): Result<String> {
        delay(1500)
        return Result.success("mock_final_banking_session_token_ios")
    }
}
