package com.shield.shared.scanner

data class AppInfo(
    val appName: String,
    val packageName: String,
    val versionName: String,
    val installerPackageName: String?,
    val isSideloaded: Boolean,
    val isThreat: Boolean,
    val certificateHash: String?,
    val threatReasons: List<String>,
    val firstInstallTime: Long,
    val lastUpdateTime: Long
)
