package org.example.shield.gate

import androidx.biometric.BiometricPrompt
import androidx.core.content.ContextCompat
import androidx.fragment.app.FragmentActivity
import kotlinx.coroutines.suspendCancellableCoroutine
import org.example.shield.AndroidContextProvider
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException

object BiometricHelper {
    suspend fun promptBiometric(): Boolean = suspendCancellableCoroutine { continuation ->
        val activity = AndroidContextProvider.context as? FragmentActivity
        if (activity == null) {
            continuation.resumeWithException(Exception("Android Context is not a FragmentActivity. Cannot show Biometric prompt."))
            return@suspendCancellableCoroutine
        }

        val executor = ContextCompat.getMainExecutor(activity)
        val biometricPrompt = BiometricPrompt(activity, executor,
            object : BiometricPrompt.AuthenticationCallback() {
                override fun onAuthenticationError(errorCode: Int, errString: CharSequence) {
                    super.onAuthenticationError(errorCode, errString)
                    continuation.resume(false)
                }

                override fun onAuthenticationSucceeded(result: BiometricPrompt.AuthenticationResult) {
                    super.onAuthenticationSucceeded(result)
                    continuation.resume(true)
                }

                override fun onAuthenticationFailed() {
                    super.onAuthenticationFailed()
                    // Called when a biometric is valid but not recognized. We wait for further attempts or error.
                }
            })

        val promptInfo = BiometricPrompt.PromptInfo.Builder()
            .setTitle("SHIELD Gate 3 Login")
            .setSubtitle("Confirm your identity to authorize transaction")
            // Allows fallback to device PIN/Password/Pattern
            .setAllowedAuthenticators(androidx.biometric.BiometricManager.Authenticators.BIOMETRIC_STRONG or androidx.biometric.BiometricManager.Authenticators.DEVICE_CREDENTIAL)
            .build()

        biometricPrompt.authenticate(promptInfo)

        continuation.invokeOnCancellation {
            biometricPrompt.cancelAuthentication()
        }
    }
}
