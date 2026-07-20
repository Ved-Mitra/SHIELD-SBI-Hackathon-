package org.example.shield

import androidx.compose.runtime.mutableStateOf

object AppLanguages {
    val availableLanguages = listOf("Bengali","Gujarati","Hindi","Kannada","Malayalam","Marathi","Odia","Punjabi","Tamil","Telugu")
    val selectedLanguage = mutableStateOf("Hindi")
    var saveLanguagePreference: ((String) -> Unit)? = null
}