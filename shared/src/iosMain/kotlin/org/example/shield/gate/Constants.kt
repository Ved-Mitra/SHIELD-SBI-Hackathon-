package org.example.shield.gate

actual object Constants {
    // iOS simulator accesses host localhost directly via 127.0.0.1 or localhost
    private const val BASE_URL = "http://127.0.0.1"
    
    actual val GATE1_URL = "$BASE_URL:8081/gate1/attest"
    actual val GATE2_URL = "$BASE_URL:8080/gate2/token"
    actual val GATE3_BEGIN_URL = "$BASE_URL:8082/gate3/authenticate/begin"
    actual val GATE3_FINISH_URL = "$BASE_URL:8082/gate3/authenticate/finish"
    actual val GATE3_REGISTER_BEGIN_URL = "$BASE_URL:8082/gate3/register/begin"
    actual val GATE3_REGISTER_FINISH_URL = "$BASE_URL:8082/gate3/register/finish"
}
