package server

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

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
	mux.Handle("/metrics", promhttp.Handler())

	//report check
	mux.HandleFunc("/report", handler.HandleReportPhishing)

	var h http.Handler = mux

	return h,nil
}

func DefaultServer(addr string, h http.Handler) *http.Server{
	return &http.Server{
		Addr: addr,
		Handler: h,
		ReadTimeout: 5*time.Second,
		WriteTimeout: 10*time.Second,
		IdleTimeout: 60*time.Second,
	}
}