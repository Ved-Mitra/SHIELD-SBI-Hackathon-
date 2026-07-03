package kafka

import (
	"context"
	"encoding/json"

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
	DeviceId string `json:"device_id"`
	Url string `json:"url"`
	Timestamp int`json:"timestamp"`
}

func PublishPhishingEvent(event PhishingEvent) error{
	if writer!=nil {
		payload, _ :=json.Marshal(event);
		return writer.WriteMessages(context.Background(),kafka.Message{Value: payload});
	}
	return nil
}

func CloseProducer() error{
	if writer!=nil{
		return writer.Close()
	}
	return nil
}