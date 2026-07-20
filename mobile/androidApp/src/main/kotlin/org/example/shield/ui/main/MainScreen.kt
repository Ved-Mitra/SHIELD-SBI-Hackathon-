package org.example.shield.ui.main

import android.content.Intent
import android.net.Uri
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.navigation3.runtime.NavKey
import org.example.shield.scanner.AppInfo
import org.example.shield.theme.ShieldTheme
import androidx.compose.ui.platform.LocalLifecycleOwner
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import android.provider.Settings
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import android.Manifest
import android.os.Build
import android.content.pm.PackageManager
import androidx.core.content.ContextCompat

@Composable
fun MainScreen(
  onItemClick: (NavKey) -> Unit,
  modifier: Modifier = Modifier,
  viewModel: MainScreenViewModel = viewModel()
) {
  val state by viewModel.uiState.collectAsStateWithLifecycle()
  val isBackgroundEnabled by viewModel.isBackgroundScanEnabled.collectAsStateWithLifecycle()
  val context = LocalContext.current

  val notificationPermissionLauncher = rememberLauncherForActivityResult(
      contract = ActivityResultContracts.RequestPermission()
  ) { isGranted ->
      // Nothing special needed, just required for the scanner to fire alerts
  }

  LaunchedEffect(Unit) {
      if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
          if (ContextCompat.checkSelfPermission(context, Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED) {
              notificationPermissionLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
          }
      }
  }

  Column(modifier = modifier.fillMaxSize()) {
    HeaderSection()

    Spacer(modifier = Modifier.height(16.dp))

    BackgroundControlCard(
      isEnabled = isBackgroundEnabled,
      onToggle = { viewModel.toggleBackgroundScan(it) },
      onRefresh = { viewModel.startScan() }
    )

    Spacer(modifier = Modifier.height(16.dp))

    when (val currentState = state) {
      MainScreenUiState.Loading -> {
        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
          CircularProgressIndicator()
        }
      }
      is MainScreenUiState.Error -> {
        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
          Text(
            text = "Error loading scan data: ${currentState.throwable.message}",
            color = MaterialTheme.colorScheme.error,
            style = MaterialTheme.typography.bodyLarge
          )
        }
      }
      is MainScreenUiState.Success -> {
        AppListSection(apps = currentState.apps)
      }
    }
  }
}

@Composable
fun HeaderSection() {
  Column(modifier = Modifier.fillMaxWidth()) {
    Text(
      text = "SHIELD Guardian",
      style = MaterialTheme.typography.headlineLarge.copy(fontWeight = FontWeight.Bold),
      color = MaterialTheme.colorScheme.primary
    )
    Text(
      text = "On-Device Passive APK Scanner & Threat Engine",
      style = MaterialTheme.typography.bodyMedium,
      color = MaterialTheme.colorScheme.onSurfaceVariant
    )
  }
}

@Composable
fun BackgroundControlCard(
  isEnabled: Boolean,
  onToggle: (Boolean) -> Unit,
  onRefresh: () -> Unit
) {
  Card(
    modifier = Modifier.fillMaxWidth(),
    shape = RoundedCornerShape(16.dp),
    colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)
  ) {
    Column(modifier = Modifier.padding(16.dp)) {
      Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween
      ) {
        Column {
          Text(
            text = "Passive Background Scan",
            style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.SemiBold),
            color = MaterialTheme.colorScheme.onSurface
          )
          Text(
            text = if (isEnabled) "Active (Scans every 15 mins via WorkManager)" else "Background service disabled",
            style = MaterialTheme.typography.bodySmall,
            color = if (isEnabled) Color(0xFF2E7D32) else MaterialTheme.colorScheme.error
          )
        }
        Switch(
          checked = isEnabled,
          onCheckedChange = onToggle
        )
      }

      Spacer(modifier = Modifier.height(12.dp))

      Button(
        onClick = onRefresh,
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(10.dp)
      ) {
        Icon(imageVector = Icons.Default.Refresh, contentDescription = "Refresh Scan")
        Spacer(modifier = Modifier.width(8.dp))
        Text(text = "Trigger Manual Scan Now")
      }
    }
  }
}

@Composable
fun AppListSection(apps: List<AppInfo>) {
  val threats = apps.filter { it.isThreat }
  val sideloaded = apps.filter { it.isSideloaded && !it.isThreat }

  Column(modifier = Modifier.fillMaxSize()) {
    Row(
      modifier = Modifier.fillMaxWidth(),
      horizontalArrangement = Arrangement.spacedBy(8.dp)
    ) {
      SummaryCard(
        title = "Scanned",
        count = apps.size,
        color = MaterialTheme.colorScheme.primary,
        modifier = Modifier.weight(1f)
      )
      SummaryCard(
        title = "Sideloaded",
        count = sideloaded.size,
        color = Color(0xFFE65100),
        modifier = Modifier.weight(1f)
      )
      SummaryCard(
        title = "Threats",
        count = threats.size,
        color = MaterialTheme.colorScheme.error,
        modifier = Modifier.weight(1f)
      )
    }

    Spacer(modifier = Modifier.height(16.dp))

    Text(
      text = "App Intelligence & Verdicts",
      style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
      color = MaterialTheme.colorScheme.onSurface
    )

    Spacer(modifier = Modifier.height(8.dp))

    LazyColumn(
      modifier = Modifier.fillMaxSize(),
      verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
      items(apps, key = { it.packageName }) { app ->
        AppInfoItem(app = app)
      }
    }
  }
}

@Composable
fun SummaryCard(title: String, count: Int, color: Color, modifier: Modifier = Modifier) {
  Card(
    modifier = modifier,
    shape = RoundedCornerShape(12.dp),
    colors = CardDefaults.cardColors(containerColor = color.copy(alpha = 0.1f))
  ) {
    Column(
      modifier = Modifier.padding(12.dp),
      horizontalAlignment = Alignment.CenterHorizontally
    ) {
      Text(
        text = count.toString(),
        style = MaterialTheme.typography.headlineMedium.copy(fontWeight = FontWeight.Bold),
        color = color
      )
      Text(
        text = title,
        style = MaterialTheme.typography.bodySmall,
        color = MaterialTheme.colorScheme.onSurface
      )
    }
  }
}

@Composable
fun AppInfoItem(app: AppInfo) {
  val context = LocalContext.current
  var expanded by remember { mutableStateOf(app.isThreat) }

  val containerColor = if (app.isThreat) {
    MaterialTheme.colorScheme.errorContainer.copy(alpha = 0.4f)
  } else if (app.isSideloaded) {
    Color(0xFFFFF3E0)
  } else {
    MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f)
  }

  Card(
    modifier = Modifier.fillMaxWidth(),
    shape = RoundedCornerShape(16.dp),
    colors = CardDefaults.cardColors(containerColor = containerColor),
    onClick = { expanded = !expanded }
  ) {
    Column(modifier = Modifier.padding(16.dp)) {
      Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween
      ) {
        Column(modifier = Modifier.weight(1f)) {
          Text(
            text = app.appName,
            style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
            color = MaterialTheme.colorScheme.onSurface
          )
          Text(
            text = "${app.packageName} (v${app.versionName})",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
          )
        }

        Spacer(modifier = Modifier.width(8.dp))

        if (app.isThreat) {
          Badge(containerColor = MaterialTheme.colorScheme.error) {
            Icon(
              imageVector = Icons.Default.Warning,
              contentDescription = "Threat",
              modifier = Modifier.size(14.dp),
              tint = MaterialTheme.colorScheme.onError
            )
            Spacer(modifier = Modifier.width(4.dp))
            Text("High Risk", color = MaterialTheme.colorScheme.onError, style = MaterialTheme.typography.labelSmall)
          }
        } else if (app.isSideloaded) {
          Badge(containerColor = Color(0xFFE65100)) {
            Text("Sideloaded", color = Color.White, style = MaterialTheme.typography.labelSmall)
          }
        } else {
          Badge(containerColor = Color(0xFF2E7D32)) {
            Icon(
              imageVector = Icons.Default.CheckCircle,
              contentDescription = "Verified",
              modifier = Modifier.size(14.dp),
              tint = Color.White
            )
            Spacer(modifier = Modifier.width(4.dp))
            Text("Verified Store", color = Color.White, style = MaterialTheme.typography.labelSmall)
          }
        }
      }

      AnimatedVisibility(visible = expanded) {
        Column(modifier = Modifier.padding(top = 12.dp)) {
          Divider(color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.2f))
          Spacer(modifier = Modifier.height(8.dp))

          Text(
            text = "Installer Source: ${app.installerPackageName ?: "Unknown / Direct Package Installer"}",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurface
          )

          if (app.certificateHash != null) {
            Spacer(modifier = Modifier.height(4.dp))
            Text(
              text = "Certificate SHA-256:\n${app.certificateHash}",
              style = MaterialTheme.typography.bodySmall,
              color = MaterialTheme.colorScheme.onSurfaceVariant
            )
          }

          if (app.threatReasons.isNotEmpty()) {
            Spacer(modifier = Modifier.height(8.dp))
            Text(
              text = "Threat Analysis Rationale:",
              style = MaterialTheme.typography.labelMedium.copy(fontWeight = FontWeight.Bold),
              color = MaterialTheme.colorScheme.error
            )
            app.threatReasons.forEach { reason ->
              Text(
                text = "• $reason",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.error
              )
            }
          }

          if (app.isThreat || app.isSideloaded) {
            Spacer(modifier = Modifier.height(12.dp))
            Button(
              onClick = {
                val intent = Intent(Intent.ACTION_UNINSTALL_PACKAGE).apply {
                  data = Uri.parse("package:${app.packageName}")
                  putExtra(Intent.EXTRA_RETURN_RESULT, true)
                }
                context.startActivity(intent)
              },
              colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.error),
              modifier = Modifier.fillMaxWidth()
            ) {
              Text(text = "Uninstall Package")
            }
          }
        }
      }
    }
  }
}

@Preview(showBackground = true)
@Composable
fun MainScreenPreview() {
  ShieldTheme {
    AppListSection(
      apps = listOf(
        AppInfo(
          appName = "YONO ŝBI (Phishing Clone)",
          packageName = "com.yonoŝbi.in.apk",
          versionName = "1.0.4",
          installerPackageName = "org.telegram.messenger",
          isSideloaded = true,
          isThreat = true,
          certificateHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
          threatReasons = listOf(
            "Unicode Homograph attack detected (Devanagari-Latin lookalike 'ŝ' in yonoŝbi.in).",
            "App sideloaded via WhatsApp/Telegram forwards bypassing Play Protect.",
            "Certificate mismatch: Fake APK signature does not match official SBI root certificate."
          ),
          firstInstallTime = 123456789L,
          lastUpdateTime = 123456789L
        )
      )
    )
  }
}
