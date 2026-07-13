package org.example.shield.ui.main

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import org.example.shield.scanner.ApkScanner
import org.example.shield.scanner.AppInfo
import org.example.shield.scanner.ScannerWorker
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class MainScreenViewModel(application: Application) : AndroidViewModel(application) {

    private val prefs = application.getSharedPreferences("shield_prefs", android.content.Context.MODE_PRIVATE)
    
    private val _uiState = MutableStateFlow<MainScreenUiState>(MainScreenUiState.Loading)
    val uiState: StateFlow<MainScreenUiState> = _uiState.asStateFlow()

    private val _isBackgroundScanEnabled = MutableStateFlow(prefs.getBoolean("bg_scan_enabled", true))
    val isBackgroundScanEnabled: StateFlow<Boolean> = _isBackgroundScanEnabled.asStateFlow()

    init {
        startScan()
        // Only schedule if it was previously enabled
        if (_isBackgroundScanEnabled.value) {
            ScannerWorker.schedule(application)
        } else {
            ScannerWorker.cancel(application)
        }
    }

    fun startScan() {
        viewModelScope.launch(Dispatchers.IO) {
            _uiState.value = MainScreenUiState.Loading
            try {
                val scanner = ApkScanner(getApplication())
                val apps = scanner.scanInstalledApps(includeSyntheticDemoThreats = true)
                _uiState.value = MainScreenUiState.Success(apps)
            } catch (e: Exception) {
                _uiState.value = MainScreenUiState.Error(e)
            }
        }
    }

    fun toggleBackgroundScan(enabled: Boolean) {
        _isBackgroundScanEnabled.value = enabled
        prefs.edit().putBoolean("bg_scan_enabled", enabled).apply()
        
        if (enabled) {
            ScannerWorker.schedule(getApplication())
        } else {
            ScannerWorker.cancel(getApplication())
        }
    }
}

sealed interface MainScreenUiState {
    object Loading : MainScreenUiState
    data class Error(val throwable: Throwable) : MainScreenUiState
    data class Success(val apps: List<AppInfo>) : MainScreenUiState
}
