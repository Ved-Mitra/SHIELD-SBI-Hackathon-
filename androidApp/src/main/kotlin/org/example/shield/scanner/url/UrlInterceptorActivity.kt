package org.example.shield.scanner.url

import org.example.shield.scanner.url.UrlClassifier

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.ui.Modifier
import java.io.InputStream

class UrlInterceptorActivity : ComponentActivity() {
    private val classifier = MobileBertClassifier()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Initialize model on a background thread to avoid blocking the main thread (ANR)
        // The 23MB ONNX model takes several seconds to load from storage
        Thread { initClassifier() }.start()

        // Read URL if the activity was launched system-wide via a clicked link
        val initialUrl = intent?.data?.toString() ?: intent?.dataString

        setContent {
            MaterialTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background.copy(alpha = if (initialUrl != null) 0.0f else 1.0f) // Transparent surface if launched as overlay
                ) {
                    // Pass the captured URL and callback functions for browser redirect and dismiss actions
                    UrlRiskReportScreen(
                        classifier = classifier, 
                        initialUrl = initialUrl,
                        onProceedToBrowser = { url -> openInSystemBrowser(url) },
                        onDismiss = { dismissOverlay() }
                    )
                }
            }
        }
    }

    private fun initClassifier() {
        try {
            // Load model bytes from assets
            val modelInputStream: InputStream = assets.open("phishing_model.onnx")
            val modelBytes = modelInputStream.readBytes()
            modelInputStream.close()

            // Load vocab content from assets
            val vocabInputStream: InputStream = assets.open("vocab.txt")
            val vocabContent = vocabInputStream.bufferedReader().use { it.readText() }
            vocabInputStream.close()

            // Initialize MobileBERT ONNX classifier
            classifier.initialize(modelBytes, vocabContent)
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    /**
     * Safely redirects the URL to system browsers, excluding the SHIELD app
     * to avoid triggering a recursive redirection loop.
     */
    private fun openInSystemBrowser(url: String) {
        try {
            val uri = Uri.parse(url)
            val intent = Intent(Intent.ACTION_VIEW, uri)
            
            // Query all activities capable of viewing the link
            val resolveInfos = packageManager.queryIntentActivities(intent, 0)
            val targetIntents = mutableListOf<Intent>()
            
            for (info in resolveInfos) {
                val packageName = info.activityInfo.packageName
                // Exclude ourselves
                if (packageName != this.packageName) {
                    val targetIntent = Intent(intent).apply {
                        setPackage(packageName)
                    }
                    targetIntents.add(targetIntent)
                }
            }
            
            if (targetIntents.isNotEmpty()) {
                // Remove the first and use it as primary, add rest as extras
                val chooserIntent = Intent.createChooser(targetIntents.removeAt(0), "Open safe link in:")
                chooserIntent.putExtra(Intent.EXTRA_INITIAL_INTENTS, targetIntents.toTypedArray())
                startActivity(chooserIntent)
            } else {
                // Fallback to normal behavior if no other browsers are found
                startActivity(intent)
            }
            finish() // Close the warning overlay
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    private fun dismissOverlay() {
        finish() // Closes this activity and returns the user to the sender app (SMS/WhatsApp)
    }
}
