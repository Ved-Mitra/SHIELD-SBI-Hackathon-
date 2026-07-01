package org.example.shield.gate

actual object Constants {
    // 10.0.2.2 points to the host machine's localhost from the Android Emulator
    private const val BASE_URL = "http://10.19.6.43"
    
    actual val GATE1_URL = "$BASE_URL:8081/gate1/attest"
    actual val GATE2_URL = "$BASE_URL:8443/gate2/token" // Hitting Envoy Proxy for mTLS
    actual val GATE3_BEGIN_URL = "$BASE_URL:8082/gate3/authenticate/begin"
    actual val GATE3_FINISH_URL = "$BASE_URL:8082/gate3/authenticate/finish"
}
