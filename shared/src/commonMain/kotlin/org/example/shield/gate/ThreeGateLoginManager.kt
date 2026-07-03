package org.example.shield.gate

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

sealed class LoginState {
    object Idle : LoginState()
    object Gate1InProgress : LoginState()
    object Gate2InProgress : LoginState()
    object Gate3InProgress : LoginState()
    data class Success(val sessionToken: String) : LoginState()
    data class Error(val message: String) : LoginState()
}

class ThreeGateLoginManager {
    private val appAttestation = AppAttestation()
    private val gateClient = GateClient()

    fun performLogin(username: String): Flow<LoginState> = flow {
        emit(LoginState.Gate1InProgress)
        
        // Generate a 32-character secure random hex nonce to prevent replay attacks
        val secureNonce = (1..32).map { "0123456789abcdef"[kotlin.random.Random.nextInt(16)] }.joinToString("")

        // 1. Gate 1: App Attestation
        val attestationResult = appAttestation.verifyAppIntegrity(secureNonce)
        if (attestationResult !is AttestationResult.Success) {
            val error = (attestationResult as? AttestationResult.Failure)?.error ?: "Gate 1 Attestation Failed"
            emit(LoginState.Error(error))
            return@flow
        }
        
        val g1Result = gateClient.getGate1Token(attestationResult.token, secureNonce)
        if (g1Result.isFailure) {
            emit(LoginState.Error(g1Result.exceptionOrNull()?.message ?: "Gate 1 Network Failed"))
            return@flow
        }
        val g1Token = g1Result.getOrThrow()

        // 2. Gate 2: mTLS Channel Auth
        emit(LoginState.Gate2InProgress)
        val g2Result = gateClient.getGate2Token(g1Token)
        if (g2Result.isFailure) {
            emit(LoginState.Error(g2Result.exceptionOrNull()?.message ?: "Gate 2 Network Failed"))
            return@flow
        }
        val g2Token = g2Result.getOrThrow()

        // 3. Gate 3: WebAuthn / FIDO2
        emit(LoginState.Gate3InProgress)
        val g3Result = gateClient.authenticateGate3(g2Token, username)
        if (g3Result.isFailure) {
            emit(LoginState.Error(g3Result.exceptionOrNull()?.message ?: "Gate 3 FIDO2 Failed"))
            return@flow
        }
        val sessionToken = g3Result.getOrThrow()

        // Final Success
        emit(LoginState.Success(sessionToken))
    }

    fun performRegistration(username: String): Flow<LoginState> = flow {
        emit(LoginState.Gate1InProgress)
        
        // Generate a 32-character secure random hex nonce to prevent replay attacks
        val secureNonce = (1..32).map { "0123456789abcdef"[kotlin.random.Random.nextInt(16)] }.joinToString("")

        // 1. Gate 1: App Attestation
        val attestationResult = appAttestation.verifyAppIntegrity(secureNonce)
        if (attestationResult !is AttestationResult.Success) {
            val error = (attestationResult as? AttestationResult.Failure)?.error ?: "Gate 1 Attestation Failed"
            emit(LoginState.Error(error))
            return@flow
        }
        
        val g1Result = gateClient.getGate1Token(attestationResult.token, secureNonce)
        if (g1Result.isFailure) {
            emit(LoginState.Error(g1Result.exceptionOrNull()?.message ?: "Gate 1 Network Failed"))
            return@flow
        }
        val g1Token = g1Result.getOrThrow()

        // 2. Gate 2: mTLS Channel Auth
        emit(LoginState.Gate2InProgress)
        val g2Result = gateClient.getGate2Token(g1Token)
        if (g2Result.isFailure) {
            emit(LoginState.Error(g2Result.exceptionOrNull()?.message ?: "Gate 2 Network Failed"))
            return@flow
        }
        val g2Token = g2Result.getOrThrow()

        // 3. Gate 3: WebAuthn / FIDO2 (REGISTER instead of authenticate)
        emit(LoginState.Gate3InProgress)
        val g3Result = gateClient.registerGate3(g2Token, username)
        if (g3Result.isFailure) {
            emit(LoginState.Error(g3Result.exceptionOrNull()?.message ?: "Gate 3 Registration Failed"))
            return@flow
        }
        val sessionToken = g3Result.getOrThrow()

        // Final Success
        emit(LoginState.Success(sessionToken))
    }
}
