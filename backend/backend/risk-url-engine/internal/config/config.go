package config

import "os"

type Config struct{
	//Addr is the TCP address the HTTP server binds to 
	Addr string 

	// kafka broker url
	KafkaBrokerUrl string
}

func Load() Config{
	return Config{
		Addr: getEnv("RISK_URL_ENGINE_ADDR","localhost:8083"),
		KafkaBrokerUrl: getEnv("KAFKA_BROKER_URL","localhost:9092"),
	}
}

func getEnv(key, fallback string) string{
	if v:=os.Getenv(key); v!="" {
		return v
	}
	return fallback
}