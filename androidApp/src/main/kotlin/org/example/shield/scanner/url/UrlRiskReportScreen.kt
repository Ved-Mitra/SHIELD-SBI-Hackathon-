package org.example.shield.scanner.url

import org.example.shield.scanner.url.RiskReport
import org.example.shield.scanner.url.RiskLevel
import org.example.shield.scanner.url.UrlRiskScorer
import org.example.shield.scanner.url.ShapFeature

import androidx.compose.animation.*
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalSoftwareKeyboardController
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.coroutines.delay

// Premium Colors matching YONO & SHIELD styling
val ThemeNavyDark = Color(0xFF0A1128)
val ThemeNavyLight = Color(0xFF101F42)
val ThemeGold = Color(0xFFD4AF37)
val ThemeSbiBlue = Color(0xFF0084C4)
val ColorSafe = Color(0xFF4CAF50)
val ColorWarning = Color(0xFFFFB300)
val ColorDanger = Color(0xFFF44336)
val ColorCritical = Color(0xFFB71C1C)

@OptIn(ExperimentalMaterial3Api::class, ExperimentalComposeUiApi::class)
@Composable
fun UrlRiskReportScreen(
    classifier: MobileBertClassifier,
    initialUrl: String? = null,
    onProceedToBrowser: (String) -> Unit = {},
    onDismiss: () -> Unit = {}
) {
    var urlInput by remember { mutableStateOf(initialUrl ?: "") }
    var riskReport by remember { mutableStateOf<RiskReport?>(null) }
    var isScanning by remember { mutableStateOf(false) }
    
    val keyboardController = LocalSoftwareKeyboardController.current
    val scrollState = rememberScrollState()

    fun performScan() {
        if (urlInput.isBlank()) return
        isScanning = true
        keyboardController?.hide()

        // Run ONNX inference on a background thread to avoid ANR
        val urlToScan = urlInput
        Thread {
            val result = UrlRiskScorer.score(urlToScan, classifier)
            // Post result back to the main thread for UI update
            android.os.Handler(android.os.Looper.getMainLooper()).post {
                riskReport = result
                isScanning = false
            }
        }.start()
    }

    // Auto-scan if launched with an initial URL from link interception (SMS / WhatsApp clicks)
    LaunchedEffect(initialUrl) {
        if (!initialUrl.isNullOrBlank()) {
            urlInput = initialUrl
            isScanning = true
            delay(400) // Brief animation delay
            // Run ONNX inference on IO dispatcher to avoid blocking the main thread
            val result = kotlinx.coroutines.withContext(kotlinx.coroutines.Dispatchers.Default) {
                UrlRiskScorer.score(initialUrl, classifier)
            }
            riskReport = result
            isScanning = false
        }
    }

    if (initialUrl != null) {
        // Overlay View Mode: Darkened backdrop with popup warning card at the bottom
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black.copy(alpha = 0.6f)),
            contentAlignment = Alignment.BottomCenter
        ) {
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .fillMaxHeight(0.85f),
                shape = RoundedCornerShape(topStart = 24.dp, topEnd = 24.dp),
                colors = CardDefaults.cardColors(containerColor = ThemeNavyLight),
                elevation = CardDefaults.cardElevation(defaultElevation = 16.dp)
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(20.dp)
                        .verticalScroll(scrollState),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    // HUD Banner Header
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(
                                color = ColorCritical.copy(alpha = 0.15f),
                                shape = RoundedCornerShape(8.dp)
                            )
                            .border(1.dp, ColorCritical.copy(alpha = 0.4f), RoundedCornerShape(8.dp))
                            .padding(12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Box(
                            modifier = Modifier
                                .size(8.dp)
                                .clip(RoundedCornerShape(4.dp))
                                .background(ColorCritical)
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = "SHIELD LINK INTERCEPT ACTIVE",
                            color = ColorCritical,
                            fontSize = 11.sp,
                            fontWeight = FontWeight.Bold,
                            letterSpacing = 1.sp
                        )
                    }

                    Spacer(modifier = Modifier.height(16.dp))

                    if (isScanning || riskReport == null) {
                        // Scan Loader
                        Column(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(160.dp),
                            verticalArrangement = Arrangement.Center,
                            horizontalAlignment = Alignment.CenterHorizontally
                        ) {
                            CircularProgressIndicator(color = ThemeSbiBlue)
                            Spacer(modifier = Modifier.height(16.dp))
                            Text("SHIELD Scanning Intercepted URL...", color = ThemeGold, fontSize = 13.sp)
                        }
                    } else {
                        // Scan Results Panel
                        val report = riskReport!!
                        val statusColor = when (report.riskLevel) {
                            RiskLevel.SAFE -> ColorSafe
                            RiskLevel.LOW -> ThemeSbiBlue
                            RiskLevel.MEDIUM -> ColorWarning
                            RiskLevel.HIGH -> ColorDanger
                            RiskLevel.CRITICAL -> ColorCritical
                        }

                        // Risk Verdict Card
                        Card(
                            modifier = Modifier
                                .fillMaxWidth()
                                .border(1.dp, statusColor.copy(alpha = 0.5f), RoundedCornerShape(16.dp)),
                            colors = CardDefaults.cardColors(containerColor = Color(0xFF162752)),
                            shape = RoundedCornerShape(16.dp)
                        ) {
                            Column(
                                modifier = Modifier.padding(16.dp),
                                horizontalAlignment = Alignment.CenterHorizontally
                            ) {
                                Text(
                                    text = "PHISHING PROBABILITY",
                                    color = Color.Gray,
                                    fontSize = 10.sp,
                                    fontWeight = FontWeight.Bold,
                                    letterSpacing = 1.sp
                                )
                                Text(
                                    text = "${report.riskPercentage}%",
                                    fontSize = 48.sp,
                                    fontWeight = FontWeight.Black,
                                    color = statusColor
                                )
                                Text(
                                    text = report.riskLevel.name,
                                    fontSize = 16.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = statusColor,
                                    letterSpacing = 1.5.sp
                                )
                                Spacer(modifier = Modifier.height(8.dp))
                                Text(
                                    text = if (report.isSafe) "This URL is classified as safe." else "WARNING: High threat level of banking credential theft!",
                                    color = Color.White,
                                    fontSize = 12.sp,
                                    textAlign = TextAlign.Center
                                )
                            }
                        }

                        Spacer(modifier = Modifier.height(12.dp))

                        // Action Buttons
                        Button(
                            onClick = { onDismiss() },
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(46.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = if (report.isSafe) ColorWarning else ColorSafe),
                            shape = RoundedCornerShape(8.dp)
                        ) {
                            Text(
                                text = if (report.isSafe) "Dismiss Warning" else "BLOCK & GO BACK (Recommended)",
                                color = Color.White,
                                fontWeight = FontWeight.Bold,
                                fontSize = 13.sp
                            )
                        }

                        Spacer(modifier = Modifier.height(8.dp))

                        TextButton(
                            onClick = { onProceedToBrowser(report.url) },
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text(
                                text = "Proceed to Browser anyway",
                                color = if (report.isSafe) Color.Gray else ColorCritical,
                                fontSize = 11.sp,
                                fontWeight = FontWeight.Bold
                            )
                        }

                        Spacer(modifier = Modifier.height(12.dp))

                        // Homograph Detection details
                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            colors = CardDefaults.cardColors(containerColor = Color(0xFF162752)),
                            shape = RoundedCornerShape(12.dp)
                        ) {
                            Column(modifier = Modifier.padding(14.dp)) {
                                Text(
                                    text = "Devanagari Homograph Audit",
                                    color = Color.White,
                                    fontSize = 14.sp,
                                    fontWeight = FontWeight.Bold
                                )
                                Spacer(modifier = Modifier.height(8.dp))
                                if (report.homographReport.isHomograph) {
                                    Text(
                                        text = "Spoofed Target: ${report.homographReport.spoofedTargetDomain}",
                                        color = ThemeGold,
                                        fontWeight = FontWeight.Bold,
                                        fontFamily = FontFamily.Monospace,
                                        fontSize = 12.sp
                                    )
                                    report.homographReport.homoglyphsFound.forEach { glyph ->
                                        Text(
                                            text = "• '${glyph.char}' mimics '${glyph.resemblesChar}'",
                                            color = Color.LightGray,
                                            fontSize = 11.sp
                                        )
                                    }
                                } else {
                                    Text(
                                        text = "No lookalikes found. Scripts: ${report.homographReport.detectedScripts.joinToString()}",
                                        color = Color.LightGray,
                                        fontSize = 11.sp
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }
    } else {
        // Normal Full Screen Dashboard Mode
        Column(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    brush = Brush.verticalGradient(
                        colors = listOf(ThemeNavyDark, ThemeNavyLight)
                    )
                )
                .padding(16.dp)
                .verticalScroll(scrollState),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            // App Title
            Spacer(modifier = Modifier.height(24.dp))
            Text(
                text = "SHIELD",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                color = Color.White,
                letterSpacing = 4.sp
            )
            Text(
                text = "URL Risk Engine • Mobile Security Portal",
                fontSize = 12.sp,
                color = ThemeGold,
                fontWeight = FontWeight.Medium,
                letterSpacing = 1.sp
            )
            Spacer(modifier = Modifier.height(24.dp))

            // Input Card
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(containerColor = Color(0xFF162752)),
                shape = RoundedCornerShape(16.dp)
            ) {
                Column(
                    modifier = Modifier.padding(16.dp),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Text(
                        text = "Scan Suspect URL",
                        color = Color.White,
                        fontSize = 16.sp,
                        fontWeight = FontWeight.SemiBold,
                        modifier = Modifier.fillMaxWidth()
                    )
                    Spacer(modifier = Modifier.height(12.dp))
                    
                    OutlinedTextField(
                        value = urlInput,
                        onValueChange = { urlInput = it },
                        placeholder = { Text("Paste URL here (e.g. http://sb०.in)", color = Color.Gray) },
                        modifier = Modifier.fillMaxWidth(),
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedTextColor = Color.White,
                            unfocusedTextColor = Color.White,
                            focusedBorderColor = ThemeSbiBlue,
                            unfocusedBorderColor = Color(0xFF2A3D6E),
                            focusedContainerColor = Color(0xFF0F1B3A),
                            unfocusedContainerColor = Color(0xFF0F1B3A)
                        ),
                        singleLine = true,
                        keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
                        keyboardActions = KeyboardActions(onSearch = { performScan() })
                    )
                    
                    Spacer(modifier = Modifier.height(16.dp))
                    
                    Button(
                        onClick = { performScan() },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(48.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = ThemeSbiBlue),
                        shape = RoundedCornerShape(8.dp),
                        enabled = urlInput.isNotBlank() && !isScanning
                    ) {
                        if (isScanning) {
                            CircularProgressIndicator(color = Color.White, modifier = Modifier.size(24.dp))
                        } else {
                            Text("Analyze URL", color = Color.White, fontWeight = FontWeight.Bold)
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Results Section
            AnimatedVisibility(
                visible = riskReport != null && !isScanning,
                enter = fadeIn() + expandVertically(),
                exit = fadeOut() + shrinkVertically()
            ) {
                riskReport?.let { report ->
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        
                        val statusColor = when (report.riskLevel) {
                            RiskLevel.SAFE -> ColorSafe
                            RiskLevel.LOW -> ThemeSbiBlue
                            RiskLevel.MEDIUM -> ColorWarning
                            RiskLevel.HIGH -> ColorDanger
                            RiskLevel.CRITICAL -> ColorCritical
                        }

                        Card(
                            modifier = Modifier
                                .fillMaxWidth()
                                .border(1.dp, statusColor.copy(alpha = 0.5f), RoundedCornerShape(16.dp)),
                            colors = CardDefaults.cardColors(containerColor = Color(0xFF162752)),
                            shape = RoundedCornerShape(16.dp)
                        ) {
                            Column(
                                modifier = Modifier.padding(20.dp),
                                horizontalAlignment = Alignment.CenterHorizontally
                            ) {
                                Text(
                                    text = "RISK ANALYSIS VERDICT",
                                    color = Color.Gray,
                                    fontSize = 11.sp,
                                    fontWeight = FontWeight.Bold,
                                    letterSpacing = 1.sp
                                )
                                Spacer(modifier = Modifier.height(8.dp))
                                
                                Text(
                                    text = "${report.riskPercentage}%",
                                    fontSize = 54.sp,
                                    fontWeight = FontWeight.Black,
                                    color = statusColor
                                )
                                
                                Text(
                                    text = report.riskLevel.name,
                                    fontSize = 18.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = statusColor,
                                    letterSpacing = 2.sp
                                )
                                
                                Spacer(modifier = Modifier.height(8.dp))
                                
                                Text(
                                    text = if (report.isSafe) "This URL is classified as relatively safe." else "WARNING: High probability of phishing or credential theft!",
                                    color = Color.White,
                                    fontSize = 13.sp,
                                    textAlign = TextAlign.Center
                                )
                            }
                        }

                        Spacer(modifier = Modifier.height(16.dp))

                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            colors = CardDefaults.cardColors(containerColor = Color(0xFF162752)),
                            shape = RoundedCornerShape(16.dp)
                        ) {
                            Column(modifier = Modifier.padding(16.dp)) {
                                Row(verticalAlignment = Alignment.CenterVertically) {
                                    Icon(
                                        imageVector = if (report.homographReport.isHomograph) Icons.Default.Warning else Icons.Default.CheckCircle,
                                        contentDescription = null,
                                        tint = if (report.homographReport.isHomograph) ColorDanger else ColorSafe,
                                        modifier = Modifier.size(24.dp)
                                    )
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text(
                                        text = "Devanagari Homograph Scan",
                                        color = Color.White,
                                        fontSize = 16.sp,
                                        fontWeight = FontWeight.Bold
                                    )
                                }
                                
                                Spacer(modifier = Modifier.height(12.dp))
                                
                                if (report.homographReport.isHomograph) {
                                    Text(
                                        text = "Devanagari-Latin lookalike characters detected! The domain is hiding non-ASCII characters to spoof a bank brand.",
                                        color = ColorDanger,
                                        fontSize = 13.sp
                                    )
                                    Spacer(modifier = Modifier.height(12.dp))
                                    
                                    Text(
                                        text = "Original Domain: ${report.homographReport.originalDomain}",
                                        color = Color.White,
                                        fontFamily = FontFamily.Monospace,
                                        fontSize = 13.sp
                                    )
                                    Text(
                                        text = "Spoofed Target: ${report.homographReport.spoofedTargetDomain}",
                                        color = ThemeGold,
                                        fontWeight = FontWeight.Bold,
                                        fontFamily = FontFamily.Monospace,
                                        fontSize = 13.sp
                                    )
                                    
                                    Spacer(modifier = Modifier.height(12.dp))
                                    Text(
                                        text = "Character Mapping Details:",
                                        color = Color.White,
                                        fontSize = 12.sp,
                                        fontWeight = FontWeight.Bold
                                    )
                                    
                                    report.homographReport.homoglyphsFound.forEach { glyph ->
                                        Text(
                                            text = "• '${glyph.char}' (${glyph.codePoint} - ${glyph.charName}) mimics Latin '${glyph.resemblesChar}'",
                                            color = Color.LightGray,
                                            fontSize = 12.sp,
                                            modifier = Modifier.padding(vertical = 2.dp)
                                        )
                                    }
                                } else {
                                    Text(
                                        text = "No Devanagari or lookalike homographs detected in the domain name. Scripts: ${report.homographReport.detectedScripts.joinToString()}",
                                        color = Color.LightGray,
                                        fontSize = 13.sp
                                    )
                                }
                            }
                        }

                        Spacer(modifier = Modifier.height(16.dp))

                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            colors = CardDefaults.cardColors(containerColor = Color(0xFF162752)),
                            shape = RoundedCornerShape(16.dp)
                        ) {
                            Column(modifier = Modifier.padding(16.dp)) {
                                Text(
                                    text = "Model Interpretability (SHAP Explanations)",
                                    color = Color.White,
                                    fontSize = 16.sp,
                                    fontWeight = FontWeight.Bold
                                )
                                Text(
                                    text = "Mathematical breakdown of how model features shifted the baseline risk (10% prior)",
                                    color = Color.Gray,
                                    fontSize = 11.sp
                                )
                                Spacer(modifier = Modifier.height(16.dp))

                                report.explanation.features.forEach { feature ->
                                    Column(modifier = Modifier.padding(vertical = 6.dp)) {
                                        Row(
                                            modifier = Modifier.fillMaxWidth(),
                                            horizontalArrangement = Arrangement.SpaceBetween,
                                            verticalAlignment = Alignment.CenterVertically
                                        ) {
                                            Text(
                                                text = feature.name,
                                                color = Color.White,
                                                fontWeight = FontWeight.SemiBold,
                                                fontSize = 13.sp,
                                                modifier = Modifier.weight(1f)
                                            )
                                            
                                            val valPrefix = if (feature.shapValue > 0) "+" else ""
                                            val displayValue = String.format("%.2f", feature.shapValue)
                                            Text(
                                                text = "$valPrefix$displayValue logit",
                                                color = if (feature.isRiskFactor) ColorDanger else ColorSafe,
                                                fontWeight = FontWeight.Bold,
                                                fontSize = 13.sp
                                            )
                                        }
                                        Text(
                                            text = feature.description,
                                            color = Color.LightGray,
                                            fontSize = 11.sp
                                        )
                                        
                                        Spacer(modifier = Modifier.height(4.dp))
                                        
                                        val logitAbs = kotlin.math.abs(feature.shapValue).toFloat()
                                        val progressFraction = (logitAbs / 4.5f).coerceIn(0.1f, 1f)
                                        
                                        Box(
                                            modifier = Modifier
                                                .fillMaxWidth()
                                                .height(4.dp)
                                                .clip(RoundedCornerShape(2.dp))
                                                .background(Color(0xFF0F1B3A))
                                        ) {
                                            Box(
                                                modifier = Modifier
                                                    .fillMaxWidth(progressFraction)
                                                    .fillMaxHeight()
                                                    .background(if (feature.isRiskFactor) ColorDanger else ColorSafe)
                                            )
                                        }
                                    }
                                }
                                
                                if (report.mlClassifierScore != null) {
                                    Spacer(modifier = Modifier.height(8.dp))
                                    Divider(color = Color(0xFF2A3D6E))
                                    Spacer(modifier = Modifier.height(8.dp))
                                    Text(
                                        text = "MobileBERT ONNX Classifier Direct Score: " + String.format("%.2f%%", (report.mlClassifierScore ?: 0.0) * 100),
                                        color = ThemeGold,
                                        fontSize = 12.sp,
                                        fontWeight = FontWeight.Bold
                                    )
                                }
                            }
                        }
                    }
                }
            }
            
            Spacer(modifier = Modifier.height(40.dp))
        }
    }
}
