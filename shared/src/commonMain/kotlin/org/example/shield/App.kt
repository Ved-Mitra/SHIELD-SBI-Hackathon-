package org.example.shield

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.safeContentPadding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.foundation.layout.padding
import androidx.compose.ui.tooling.preview.Preview
import kotlinx.coroutines.launch
import org.example.shield.gate.AppAttestation
import org.example.shield.gate.AttestationResult

@Composable
@Preview
fun App() {
    MaterialTheme {
        var attestationStatus by remember { mutableStateOf("Verifying App Authenticity (Gate 1)...") }
        val scope = rememberCoroutineScope()

        LaunchedEffect(Unit) {
            scope.launch {
                val result = AppAttestation().verifyAppIntegrity()
                attestationStatus = when (result) {
                    is AttestationResult.Success -> "Gate 1 Passed: Hardware Attested ✓"
                    is AttestationResult.Failure -> "Gate 1 Failed: ${result.error}"
                }
            }
        }

        Column(
            modifier = Modifier
                .background(MaterialTheme.colorScheme.background)
                .safeContentPadding()
                .fillMaxSize(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = androidx.compose.foundation.layout.Arrangement.Center
        ) {
            Text(
                "This is a demo app based on SHIELD",
                style = MaterialTheme.typography.titleLarge
            )
            Text(
                text = attestationStatus,
                style = MaterialTheme.typography.bodyMedium,
                modifier = Modifier.padding(top = 16.dp)
            )
        }
    }
}