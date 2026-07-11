package org.example.shield.scanner.url

class HomoglyphInfo(
    val char: Char,
    val codePoint: String,
    val charName: String,
    val resemblesChar: String
)

class HomographReport(
    val isHomograph: Boolean,
    val isMixedScript: Boolean,
    val originalDomain: String,
    val spoofedTargetDomain: String,
    val detectedScripts: List<String>,
    val homoglyphsFound: List<HomoglyphInfo>
)

object DevanagariHomographDetector {
    // Mapping of Devanagari characters to their ASCII lookalikes
    private val DEVANAGARI_LOOKALIKES = mapOf(
        '०' to "o", // Devanagari Digit Zero -> 'o'
        '१' to "1", // Devanagari Digit One -> '1' or 'l'
        '२' to "2", // Devanagari Digit Two -> '2'
        '३' to "3", // Devanagari Digit Three -> '3'
        '७' to "7", // Devanagari Digit Seven -> '7'
        '८' to "8", // Devanagari Digit Eight -> '8'
        'ा' to "l", // Devanagari Vowel Sign Aa -> 'l'
        'ी' to "l", // Devanagari Vowel Sign I -> 'l'
        'ो' to "l", // Devanagari Vowel Sign O -> 'l' or 'f'
        '।' to "l"  // Devanagari Danda -> 'l'
    )
    
    // Accented/Extended Latin and Cyrillic lookalikes
    private val OTHER_LOOKALIKES = mapOf(
        'ŝ' to "s",
        'а' to "a", // Cyrillic a
        'е' to "e", // Cyrillic e
        'о' to "o", // Cyrillic o
        'р' to "p", // Cyrillic p
        'с' to "c", // Cyrillic c
        'у' to "y", // Cyrillic y
        'х' to "x", // Cyrillic x
        'і' to "i"  // Cyrillic i
    )

    fun extractDomain(url: String): String {
        var clean = url.trim()
        if (clean.startsWith("http://", ignoreCase = true)) {
            clean = clean.substring(7)
        } else if (clean.startsWith("https://", ignoreCase = true)) {
            clean = clean.substring(8)
        }
        val slashIndex = clean.indexOf('/')
        if (slashIndex != -1) {
            clean = clean.substring(0, slashIndex)
        }
        val colonIndex = clean.indexOf(':')
        if (colonIndex != -1) {
            clean = clean.substring(0, colonIndex)
        }
        return clean
    }

    fun analyze(url: String): HomographReport {
        val domain = extractDomain(url)
        val homoglyphs = mutableListOf<HomoglyphInfo>()
        val scripts = mutableSetOf<String>()
        var hasLatin = false
        var hasDevanagari = false
        var hasOther = false
        
        val spoofedDomainBuilder = StringBuilder()
        
        for (char in domain) {
            val codePoint = char.code
            val codePointHex = "U+" + codePoint.toString(16).uppercase().padStart(4, '0')
            
            when {
                // ASCII A-Z / a-z
                (codePoint in 65..90) || (codePoint in 97..122) -> {
                    hasLatin = true
                    scripts.add("Latin")
                    spoofedDomainBuilder.append(char)
                }
                // Devanagari range
                codePoint in 0x0900..0x097F -> {
                    hasDevanagari = true
                    scripts.add("Devanagari")
                    val lookalike = DEVANAGARI_LOOKALIKES[char]
                    if (lookalike != null) {
                        homoglyphs.add(
                            HomoglyphInfo(
                                char = char,
                                codePoint = codePointHex,
                                charName = getCharName(char),
                                resemblesChar = lookalike
                            )
                        )
                        spoofedDomainBuilder.append(lookalike)
                    } else {
                        spoofedDomainBuilder.append(char)
                    }
                }
                else -> {
                    // Check other lookalikes (like Cyrillic, Accented Latin)
                    val lookalike = OTHER_LOOKALIKES[char]
                    if (lookalike != null) {
                        hasOther = true
                        scripts.add("Extended/Cyrillic")
                        homoglyphs.add(
                            HomoglyphInfo(
                                char = char,
                                codePoint = codePointHex,
                                charName = getCharName(char),
                                resemblesChar = lookalike
                            )
                        )
                        spoofedDomainBuilder.append(lookalike)
                    } else {
                        if (codePoint != '.'.code && codePoint != '-'.code) {
                            scripts.add("Other")
                        }
                        spoofedDomainBuilder.append(char)
                    }
                }
            }
        }
        
        val isMixedScript = (hasLatin && (hasDevanagari || hasOther)) || (hasDevanagari && hasOther)
        val isHomograph = homoglyphs.isNotEmpty()
        
        return HomographReport(
            isHomograph = isHomograph,
            isMixedScript = isMixedScript,
            originalDomain = domain,
            spoofedTargetDomain = spoofedDomainBuilder.toString(),
            detectedScripts = scripts.toList(),
            homoglyphsFound = homoglyphs
        )
    }
    
    private fun getCharName(char: Char): String {
        return when (char) {
            '०' -> "DEVANAGARI DIGIT ZERO"
            '१' -> "DEVANAGARI DIGIT ONE"
            '२' -> "DEVANAGARI DIGIT TWO"
            '३' -> "DEVANAGARI DIGIT THREE"
            '७' -> "DEVANAGARI DIGIT SEVEN"
            '८' -> "DEVANAGARI DIGIT EIGHT"
            'ा' -> "DEVANAGARI VOWEL SIGN AA"
            'ी' -> "DEVANAGARI VOWEL SIGN I"
            'ो' -> "DEVANAGARI VOWEL SIGN O"
            '।' -> "DEVANAGARI DANDA"
            'ŝ' -> "LATIN SMALL LETTER S WITH CIRCUMFLEX"
            'а' -> "CYRILLIC SMALL LETTER A"
            'е' -> "CYRILLIC SMALL LETTER IE"
            'о' -> "CYRILLIC SMALL LETTER O"
            'р' -> "CYRILLIC SMALL LETTER ER"
            'с' -> "CYRILLIC SMALL LETTER ES"
            'у' -> "CYRILLIC SMALL LETTER U"
            'х' -> "CYRILLIC SMALL LETTER HA"
            'і' -> "CYRILLIC SMALL LETTER BYELORUSSIAN-UKRAINIAN I"
            else -> "UNICODE CHARACTER " + char.code.toString(16).uppercase()
        }
    }
}
