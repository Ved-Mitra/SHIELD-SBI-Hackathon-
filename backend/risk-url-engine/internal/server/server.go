package server

import (
	"net/http"
	"shield/risk-url-engine/internal/config"
	"shield/risk-url-engine/internal/handler"
	"shield/risk-url-engine/internal/kafka"
)


func New(cfg config.Config) (http.Handler, error) {
	//intiliazing kafka producer
	kafka.InitProducer(cfg.KafkaBrokerUrl)

	mux := http.NewServeMux()

	// health check
	mux.HandleFunc("/healthz", handler.Health)

	//report check
	mux.HandleFunc("/report", handler.HandleReportPhishing)

	var h http.Handler = mux

	return h,nil
}