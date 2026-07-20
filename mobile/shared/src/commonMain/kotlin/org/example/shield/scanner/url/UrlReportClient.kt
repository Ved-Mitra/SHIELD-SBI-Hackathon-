package org.example.shield.scanner.url

import io.ktor.client.*
import io.ktor.client.call.*
import io.ktor.client.plugins.contentnegotiation.*
import io.ktor.client.request.*
import io.ktor.http.*
import io.ktor.serialization.kotlinx.json.*
import kotlinx.serialization.json.Json
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.IO
import kotlinx.coroutines.withContext


class UrlReportClient {
    private val client = HttpClient{
        install(ContentNegotiation){
            json(Json{
                ignoreUnknownKeys=true
                prettyPrint=true
            })
        }
    }

    suspend fun reportUrl(report: UrlReportRequest): Result<UrlReportResponse> {
        return withContext(Dispatchers.IO){
            try{
                val response = client.post("http://10.19.6.43:8083/report"){
                    contentType(ContentType.Application.Json)
                    setBody(report)
                }
                if(response.status.isSuccess()){
                    val result: UrlReportResponse = response.body()
                    Result.success(result)
                }else{
                    Result.failure(Exception("Failed to report ${response.status}"))
                }
            } catch(e: Exception){
                Result.failure(e)
            }
        }
    }
}
