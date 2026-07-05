package com.shield.shared.scanner

/**
 * iOS stub for ApkScanner. APK scanning is Android-only functionality.
 * On iOS, this always returns an empty list.
 */
class ApkScanner {
    fun scanInstalledApps(includeSyntheticDemoThreats: Boolean = false): List<AppInfo> {
        return emptyList()
    }
}
