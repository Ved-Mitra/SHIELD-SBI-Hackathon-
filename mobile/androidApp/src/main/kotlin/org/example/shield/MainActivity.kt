package org.example.shield

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.Composable
import androidx.compose.ui.tooling.preview.Preview

import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.Box
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Security
import androidx.compose.material.icons.filled.Search
import org.example.shield.ui.main.MainScreen
import androidx.fragment.app.FragmentActivity

class MainActivity : FragmentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge(
            statusBarStyle = androidx.activity.SystemBarStyle.light(
                android.graphics.Color.TRANSPARENT,
                android.graphics.Color.TRANSPARENT
            )
        )
        super.onCreate(savedInstanceState)
        
        // DEV ONLY: Bypass Hostname Verification for Gate-2 mTLS IP mismatch
        javax.net.ssl.HttpsURLConnection.setDefaultHostnameVerifier { _, _ -> true }
        
        // Provide the Activity context for BiometricPrompt
        AndroidContextProvider.context = this

        // Load Language Preferences
        val prefs = getSharedPreferences("ShieldPrefs", android.content.Context.MODE_PRIVATE)
        AppLanguages.selectedLanguage.value = prefs.getString("selected_language", "Hindi") ?: "Hindi"
        AppLanguages.saveLanguagePreference = { newLang ->
            prefs.edit().putString("selected_language", newLang).apply()
        }

        setContent {
            var selectedTab by remember { mutableStateOf(0) }
            
            Scaffold(
                bottomBar = {
                    NavigationBar {
                        NavigationBarItem(
                            icon = { Icon(Icons.Filled.Security, contentDescription = "Gate Login") },
                            label = { Text("Gate Login") },
                            selected = selectedTab == 0,
                            onClick = { selectedTab = 0 }
                        )
                        NavigationBarItem(
                            icon = { Icon(Icons.Filled.Search, contentDescription = "Scanner") },
                            label = { Text("Scanner") },
                            selected = selectedTab == 1,
                            onClick = { selectedTab = 1 }
                        )
                    }
                }
            ) { paddingValues ->
                Box(modifier = Modifier.padding(paddingValues)) {
                    if (selectedTab == 0) {
                        App()
                    } else {
                        org.example.shield.ui.main.MainScreen(onItemClick = {})
                    }
                }
            }
        }
    }
}

@Preview
@Composable
fun AppAndroidPreview() {
    App()
}