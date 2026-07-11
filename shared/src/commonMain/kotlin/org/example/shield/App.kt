package org.example.shield

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.tooling.preview.Preview
import kotlinx.coroutines.launch
import org.example.shield.gate.LoginState
import org.example.shield.gate.ThreeGateLoginManager

@Composable
@Preview
fun App() {
    MaterialTheme {
        var loginState by remember { mutableStateOf<LoginState>(LoginState.Idle) }
        val scope = rememberCoroutineScope()
        val loginManager = remember { ThreeGateLoginManager() }

        Column(
            modifier = Modifier
                .background(MaterialTheme.colorScheme.background)
                .safeContentPadding()
                .fillMaxSize()
                .padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                "SHIELD - 3-Gate Authenticator",
                style = MaterialTheme.typography.headlineSmall,
                color = MaterialTheme.colorScheme.primary,
                modifier = Modifier.padding(bottom = 32.dp)
            )

            Column(modifier = Modifier.fillMaxWidth(), horizontalAlignment = Alignment.CenterHorizontally) {
                Button(
                    onClick = {
                        scope.launch {
                            loginManager.performRegistration("ved@sbi.example").collect { state ->
                                loginState = state
                            }
                        }
                    },
                    enabled = loginState == LoginState.Idle || loginState is LoginState.Error,
                    modifier = Modifier.fillMaxWidth().height(50.dp)
                ) {
                    Text("1. Register passkey")
                }

                Spacer(modifier = Modifier.height(16.dp))

                Button(
                    onClick = {
                        scope.launch {
                            loginManager.performLogin("ved@sbi.example").collect { state ->
                                loginState = state
                            }
                        }
                    },
                    enabled = loginState == LoginState.Idle || loginState is LoginState.Error,
                    modifier = Modifier.fillMaxWidth().height(50.dp)
                ) {
                    Text("2. Authenticate")
                }
            }

            Spacer(modifier = Modifier.height(32.dp))

            // Status Display
            when (val state = loginState) {
                is LoginState.Idle -> {
                    Text("Ready to authenticate.", style = MaterialTheme.typography.bodyLarge)
                }
                is LoginState.Gate1InProgress -> {
                    CircularProgressIndicator()
                    Spacer(modifier = Modifier.height(16.dp))
                    Text("Gate 1: Attesting Device & App Integrity...", style = MaterialTheme.typography.bodyLarge)
                }
                is LoginState.Gate2InProgress -> {
                    CircularProgressIndicator(color = MaterialTheme.colorScheme.secondary)
                    Spacer(modifier = Modifier.height(16.dp))
                    Text("Gate 1 ✓ Passed", color = androidx.compose.ui.graphics.Color(0xFF4CAF50))
                    Text("Gate 2: Establishing mTLS Channel...", style = MaterialTheme.typography.bodyLarge)
                }
                is LoginState.Gate3InProgress -> {
                    CircularProgressIndicator(color = MaterialTheme.colorScheme.tertiary)
                    Spacer(modifier = Modifier.height(16.dp))
                    Text("Gate 1 ✓ Passed", color = androidx.compose.ui.graphics.Color(0xFF4CAF50))
                    Text("Gate 2 ✓ Passed", color = androidx.compose.ui.graphics.Color(0xFF4CAF50))
                    Text("Gate 3: Prompting FIDO2 Biometrics...", style = MaterialTheme.typography.bodyLarge)
                }
                is LoginState.Success -> {
                    Text("Gate 1 ✓ Passed", color = androidx.compose.ui.graphics.Color(0xFF4CAF50))
                    Text("Gate 2 ✓ Passed", color = androidx.compose.ui.graphics.Color(0xFF4CAF50))
                    Text("Gate 3 ✓ Passed", color = androidx.compose.ui.graphics.Color(0xFF4CAF50))
                    Spacer(modifier = Modifier.height(24.dp))
                    Card(
                        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.primaryContainer),
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Text("🎉 Authentication Successful!", style = MaterialTheme.typography.titleMedium)
                            Text("Session Token: ${state.sessionToken.take(15)}...", style = MaterialTheme.typography.bodySmall)
                        }
                    }
                }
                is LoginState.Error -> {
                    Text("❌ Authentication Failed", color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.titleMedium)
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(state.message, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodyMedium)
                }
            }
        }
    }
}