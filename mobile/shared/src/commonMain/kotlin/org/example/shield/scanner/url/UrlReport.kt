package org.example.shield.scanner.url
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class UrlReportRequest(
    @SerialName("url") val url:String,
    @SerialName("device_id") val deviceId:String,
    @SerialName("timestamp") val timestamp:Long
)

@Serializable
data class UrlReportResponse(
    @SerialName("status") val status:String,
    @SerialName("message") val message: String
)
