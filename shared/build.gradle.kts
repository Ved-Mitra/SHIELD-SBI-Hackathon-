plugins {
    alias(libs.plugins.kotlin.multiplatform)
    alias(libs.plugins.android.kotlin.multiplatform.library)
}

kotlin {
    androidLibrary {
        namespace = "com.shield.shared"
        compileSdk = 36
        minSdk = 24
    }

    // iOS targets (stub — no iOS-specific code for this prototype)
    iosArm64()
    iosSimulatorArm64()

    sourceSets {
        commonMain.dependencies {
            implementation(libs.kotlinx.coroutines.core)
        }
        androidMain.dependencies {
            // Android SDK types like Context, PackageManager are available here
        }
    }
}
