package com.shield.shared.scanner.url

import kotlin.math.exp

class ShapFeature(
    val name: String,
    val shapValue: Double, // Contribution in logit space
    val description: String,
    val isRiskFactor: Boolean // True for risk increases, false for trust builders
)

class ShapExplanation(
    val baseProbability: Double,
    val finalProbability: Double,
    val features: List<ShapFeature>
)

object ShapExplainer {
    const val BASE_LOGIT = -2.2

    fun sigmoid(logit: Double): Double {
        return 1.0 / (1.0 + exp(-logit))
    }

    fun explain(
        url: String,
        homographReport: HomographReport,
        hasIpAddress: Boolean,
        urlLength: Int,
        subdomainCount: Int,
        suspectKeywords: List<String>,
        isHttps: Boolean,
        hasPort: Boolean,
        isSuspiciousTld: Boolean,
        isTrustedDomain: Boolean,
        isTyposquatted: Boolean
    ): ShapExplanation {
        val features = mutableListOf<ShapFeature>()
        var runningLogit = BASE_LOGIT

        if (isTrustedDomain) {
            val valLogit = -4.5
            runningLogit += valLogit
            features.add(
                ShapFeature(
                    name = "Whitelisted Domain",
                    shapValue = valLogit,
                    description = "Matches official institution domain whitelist",
                    isRiskFactor = false
                )
            )
        } else {
            if (homographReport.isHomograph) {
                val valLogit = 3.5
                runningLogit += valLogit
                features.add(
                    ShapFeature(
                        name = "Devanagari Homograph Spoofing",
                        shapValue = valLogit,
                        description = "Contains Devanagari characters lookalikes to Latin letters",
                        isRiskFactor = true
                    )
                )
            }

            if (isTyposquatted) {
                val valLogit = 3.0
                runningLogit += valLogit
                features.add(
                    ShapFeature(
                        name = "Brand Typosquatting / Lookalike",
                        shapValue = valLogit,
                        description = "Domain name contains lookalike characters or typosquatting targeting official brand names",
                        isRiskFactor = true
                    )
                )
            }

            if (homographReport.isMixedScript && !homographReport.isHomograph) {
                val valLogit = 1.5
                runningLogit += valLogit
                features.add(
                    ShapFeature(
                        name = "Mixed Unicode Scripts",
                        shapValue = valLogit,
                        description = "Mixes Latin with other Unicode scripts in domain name",
                        isRiskFactor = true
                    )
                )
            }

            if (hasIpAddress) {
                val valLogit = 2.5
                runningLogit += valLogit
                features.add(
                    ShapFeature(
                        name = "IP Address Hostname",
                        shapValue = valLogit,
                        description = "Uses raw IP address instead of registered domain name",
                        isRiskFactor = true
                    )
                )
            }

            if (suspectKeywords.isNotEmpty()) {
                val keywordLogit = suspectKeywords.size * 1.5
                runningLogit += keywordLogit
                features.add(
                    ShapFeature(
                        name = "Brand & Urgency Keywords",
                        shapValue = keywordLogit,
                        description = "Contains suspicious keywords: ${suspectKeywords.joinToString(", ")}",
                        isRiskFactor = true
                    )
                )
            }

            if (isSuspiciousTld) {
                val valLogit = 1.2
                runningLogit += valLogit
                features.add(
                    ShapFeature(
                        name = "Suspicious TLD",
                        shapValue = valLogit,
                        description = "Uses top-level domain frequently associated with spam or phishing (.xyz, .top, etc.)",
                        isRiskFactor = true
                    )
                )
            }

            if (urlLength > 75) {
                val valLogit = 0.6
                runningLogit += valLogit
                features.add(
                    ShapFeature(
                        name = "Excessive URL Length",
                        shapValue = valLogit,
                        description = "URL is abnormally long (${urlLength} characters), typical for padding/obfuscation",
                        isRiskFactor = true
                    )
                )
            }

            if (subdomainCount > 3) {
                val valLogit = 0.8
                runningLogit += valLogit
                features.add(
                    ShapFeature(
                        name = "Deep Subdomains",
                        shapValue = valLogit,
                        description = "Has ${subdomainCount} subdomains, which can mask the true root host",
                        isRiskFactor = true
                    )
                )
            }

            if (hasPort) {
                val valLogit = 0.5
                runningLogit += valLogit
                features.add(
                    ShapFeature(
                        name = "Non-Standard Port",
                        shapValue = valLogit,
                        description = "Explicitly binds connection to a custom port",
                        isRiskFactor = true
                    )
                )
            }

            if (isHttps) {
                val valLogit = -0.8
                runningLogit += valLogit
                features.add(
                    ShapFeature(
                        name = "Secure HTTPS Protocol",
                        shapValue = valLogit,
                        description = "Uses active TLS encryption",
                        isRiskFactor = false
                    )
                )
            } else {
                val valLogit = 2.0
                runningLogit += valLogit
                features.add(
                    ShapFeature(
                        name = "Insecure HTTP Protocol",
                        shapValue = valLogit,
                        description = "Transmits credentials in plain text without SSL/TLS",
                        isRiskFactor = true
                    )
                )
            }
        }

        val baseProbability = sigmoid(BASE_LOGIT)
        val finalProbability = sigmoid(runningLogit)

        return ShapExplanation(
            baseProbability = baseProbability,
            finalProbability = finalProbability,
            features = features
        )
    }
}
