package org.example.shield

import android.os.Bundle
import androidx.fragment.app.FragmentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.Composable
import androidx.compose.ui.tooling.preview.Preview

class MainActivity : FragmentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge()
        super.onCreate(savedInstanceState)
        
        // DEV ONLY: Bypass Hostname Verification for Gate-2 mTLS IP mismatch
        javax.net.ssl.HttpsURLConnection.setDefaultHostnameVerifier { _, _ -> true }
        
        // Provide the Activity context for BiometricPrompt
        AndroidContextProvider.context = this

        setContent {
            App()
        }
    }
}

@Preview
@Composable
fun AppAndroidPreview() {
    App()
}