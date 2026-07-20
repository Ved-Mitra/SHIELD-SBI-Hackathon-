package org.example.shield

interface Platform {
    val name: String
}

expect fun getPlatform(): Platform