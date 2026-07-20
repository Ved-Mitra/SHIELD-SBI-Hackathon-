package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"shield/risk-url-engine/internal/kafka"
)



type Report struct{
	Url string `json:"url"`
	DeviceId string `json:"device_id"`
	Timestamp int64 `json:"timestamp"`
}

type Response struct{
	Status string `json:"status"`
	Message string `json:"message"`
}

func HandleReportPhishing(w http.ResponseWriter, r* http.Request){
	log.Printf("[INFO] Received %s request to /report from %s", r.Method, r.RemoteAddr)
	if r.Method!=http.MethodPost {
		http.Error(w,"Wrong Method Call",http.StatusMethodNotAllowed)
		return
	}

	var payload Report
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err!=nil{
		http.Error(w, `{"status":"error","message":"Wrong Json format"}`,http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if payload.Url==""{
		http.Error(w,"Url filed cannot be empty",http.StatusBadRequest)
		return
	}

	if payload.DeviceId==""{
		http.Error(w,"DeviceId cannot be empty",http.StatusBadRequest)
		return
	}

	go kafka.PublishPhishingEvent(kafka.PhishingEvent{DeviceId: payload.DeviceId, Url: payload.Url, Timestamp: payload.Timestamp})

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)

	successResponse := Response{
		Status: "success",
		Message: "Phishing Url logged",
	}
	json.NewEncoder(w).Encode(successResponse)
}