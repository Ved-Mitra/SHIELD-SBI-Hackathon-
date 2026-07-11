package org.example.shield.scanner.url

/** Platform-provided classifier interface - implemented by MobileBertClassifier on Android. */
interface UrlClassifier {
    fun classify(url: String): Double?
}

class RiskReport(
    val url: String,
    val isSafe: Boolean,
    val riskLevel: RiskLevel,
    val riskPercentage: Int,
    val homographReport: HomographReport,
    val explanation: ShapExplanation,
    val mlClassifierScore: Double?
)

enum class RiskLevel {
    SAFE,       // 0 - 15%
    LOW,        // 16 - 35%
    MEDIUM,     // 36 - 60%
    HIGH,       // 61 - 85%
    CRITICAL    // 86 - 100%
}

object UrlRiskScorer {
    private val TRUSTED_DOMAINS = listOf(
        "sbi.co.in",
        "onlinesbi.sbi",
        "onlinesbi.com",
        "statebankofindia.com",
        "sbi"
    )

    private val SUSPICIOUS_TLDS = listOf(
        "xyz", "top", "click", "link", "site", "gq", "cf", "tk", "ml", "ga", "club", "info", "loan"
    )

    private val BRAND_KEYWORDS = listOf(
        "yono", "sbi", "statebank", "sbicard", "yonoapps", "yonosbi", "onlinesbi", "netbanking"
    )

    private val URGENCY_KEYWORDS = listOf(
        "kyc", "pan", "aadhar", "verify", "verification", "update", "blocked", "suspended",
        "login", "signin", "auth", "secure", "bonus", "reward", "points", "wallet",
        "customer-support", "helpline", "care", "service", "support", "gift", "cashback",
        "refund", "cash", "win", "lucky", "draw", "apk", "download", "pay", "upi", "gpay", "phonepe", "paytm"
    )

    fun checkIsTrusted(domain: String): Boolean {
        // Any domain ending in .sbi is owned by SBI
        if (domain.endsWith(".sbi", ignoreCase = true)) {
            return true
        }
        return TRUSTED_DOMAINS.any { trusted ->
            domain.equals(trusted, ignoreCase = true) || domain.endsWith(".$trusted", ignoreCase = true)
        }
    }

    private fun checkIsSuspiciousTld(domain: String): Boolean {
        val parts = domain.split(".")
        if (parts.size > 1) {
            val tld = parts.last().lowercase()
            return SUSPICIOUS_TLDS.contains(tld)
        }
        return false
    }

    private fun hasIpAddressHost(domain: String): Boolean {
        // Matches standard IPv4 pattern (e.g. 192.168.1.1)
        val ipv4Regex = Regex("^\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}$")
        return ipv4Regex.matches(domain)
    }

    private fun levenshteinDistance(s1: String, s2: String): Int {
        val dp = IntArray(s2.length + 1) { it }
        for (i in 1..s1.length) {
            var prev = dp[0]
            dp[0] = i
            for (j in 1..s2.length) {
                val temp = dp[j]
                if (s1[i - 1] == s2[j - 1]) {
                    dp[j] = prev
                } else {
                    dp[j] = minOf(dp[j - 1], dp[j], prev) + 1
                }
                prev = temp
            }
        }
        return dp[s2.length]
    }

    fun checkIsTyposquatted(domain: String): Boolean {
        val parts = domain.split(".", "-")
        for (part in parts) {
            val cleanPart = part.lowercase()
            // Check similarity to "sbi"
            if (cleanPart != "sbi" && cleanPart.length == 3 && levenshteinDistance(cleanPart, "sbi") == 1) {
                return true
            }
            // Check similarity to "yono"
            if (cleanPart != "yono" && cleanPart.length == 4 && levenshteinDistance(cleanPart, "yono") <= 2) {
                return true
            }
            // Check similarity to "statebank"
            if (cleanPart != "statebank" && levenshteinDistance(cleanPart, "statebank") <= 2) {
                return true
            }
        }
        return false
    }

    fun score(url: String, mlClassifier: UrlClassifier? = null): RiskReport {
        val domain = DevanagariHomographDetector.extractDomain(url)
        val homographReport = DevanagariHomographDetector.analyze(url)
        
        val isTrusted = checkIsTrusted(domain)
        val isSuspiciousTld = checkIsSuspiciousTld(domain)
        val hasIp = hasIpAddressHost(domain)
        val isHttps = url.startsWith("https://", ignoreCase = true)
        val hasPort = domain.contains(":") || (url.substringAfter("://", "").substringBefore("/", "").contains(":"))
        val isTyposquatted = checkIsTyposquatted(domain)
        
        // Find brand and urgency keywords
        val foundKeywords = mutableListOf<String>()
        val lowercaseUrl = url.lowercase()
        
        for (kw in BRAND_KEYWORDS) {
            if (lowercaseUrl.contains(kw)) {
                foundKeywords.add(kw)
            }
        }
        for (kw in URGENCY_KEYWORDS) {
            if (lowercaseUrl.contains(kw)) {
                foundKeywords.add(kw)
            }
        }

        // Subdomain count
        val domainParts = domain.split(".")
        val subdomainCount = if (domainParts.size > 2) domainParts.size - 2 else 0

        // Compute SHAP explanations
        val explanation = ShapExplainer.explain(
            url = url,
            homographReport = homographReport,
            hasIpAddress = hasIp,
            urlLength = url.length,
            subdomainCount = subdomainCount,
            suspectKeywords = foundKeywords,
            isHttps = isHttps,
            hasPort = hasPort,
            isSuspiciousTld = isSuspiciousTld,
            isTrustedDomain = isTrusted,
            isTyposquatted = isTyposquatted
        )

        // Evaluate model classification if provided
        val mlScore = mlClassifier?.classify(url)
        
        // Final probability computation: Take maximum of ML and Heuristic scores to ensure confidence peaks
        var finalProbability = if (mlScore != null && !isTrusted) {
            maxOf(mlScore, explanation.finalProbability)
        } else {
            explanation.finalProbability
        }

        // Security override: Force high/critical threat if a homograph or typosquatting is verified
        if ((homographReport.isHomograph || isTyposquatted) && !isTrusted) {
            if (finalProbability < 0.82) {
                finalProbability = 0.82 + (finalProbability * 0.15) // Boosts to 82% - 97% range
            }
        }

        val riskPercentage = (finalProbability * 100).toInt().coerceIn(0, 100)

        val riskLevel = when {
            riskPercentage <= 15 -> RiskLevel.SAFE
            riskPercentage <= 35 -> RiskLevel.LOW
            riskPercentage <= 60 -> RiskLevel.MEDIUM
            riskPercentage <= 85 -> RiskLevel.HIGH
            else -> RiskLevel.CRITICAL
        }

        // Strict safety definition: only completely SAFE (0-15%) urls are considered safe
        val isSafe = riskLevel == RiskLevel.SAFE

        return RiskReport(
            url = url,
            isSafe = isSafe,
            riskLevel = riskLevel,
            riskPercentage = riskPercentage,
            homographReport = homographReport,
            explanation = explanation,
            mlClassifierScore = mlScore
        )
    }
}
