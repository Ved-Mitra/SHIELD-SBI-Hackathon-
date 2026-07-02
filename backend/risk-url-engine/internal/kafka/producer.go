package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

func InitProducer(brokerUrl string){
	writer= &kafka.Writer{
		Addr: kafka.TCP(brokerUrl),
		Topic: "url-events",
		Balancer: &kafka.LeastBytes{},
	}
}

type PhishingEvent struct{
	deviceId string `json:"device_id"`
	Url string `json:"url"`
	Timestamp time.Time `json:"timestamp"`
}

func PublishPhishingEvent(event PhishingEvent) error{
	if writer==nil{
		fmt.Println("Kafka Writer is NULL in risk-url-engine")
	}
	payload, _ := json.Marshal(event)
	return writer.WriteMessages(context.Background(),kafka.Message{Value: payload})
}

func CloseProducer() error{
	if writer!=nil{
		return writer.Close()
	}
	return nil
}