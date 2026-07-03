package kafka

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer;

func InitProducer(brokerURL string){
	writer=&kafka.Writer{
		Addr: kafka.TCP(brokerURL),
		Topic: "auth-events",
		Balancer: &kafka.LeastBytes{},
	}
}

type AuthEvent struct{
	UserID string `json:"user_id"`
	Gate int `json:"gate"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func PublishEvent(event AuthEvent) error{
	if writer!=nil {
		payload, _ :=json.Marshal(event);
		return writer.WriteMessages(context.Background(),kafka.Message{Value: payload});
	}
	return nil
}

func CloseProducer() error{
	if writer!=nil {
		return writer.Close()
	}
	return nil
}