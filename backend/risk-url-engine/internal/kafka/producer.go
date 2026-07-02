package kafka

import (
	"context"
	"encoding/json"
	"fmt"

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

type AuthEvent struct{
	UserId string `json:"user_id"`
	Url string `json:"url"`
}

func PublishPhishingEvent(event AuthEvent) error{
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