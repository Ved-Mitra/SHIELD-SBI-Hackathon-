package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

func InitProducer(brokerURL string){
	writer=&kafka.Writer{
		Addr: kafka.TCP(brokerURL),
		Topic: "auth-events",
		Balancer: &kafka.LeastBytes{},
		BatchSize: 1, // <--- force immediate write
		BatchTimeout: 10 * time.Millisecond,
		MaxAttempts: 3, // <--- don't retry forever
		ReadTimeout: 2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
}

type AuthEvent struct{
	UserID string `json:"user_id"`
	Gate int `json:"gate"`
	Status string `json:"status"`
	Reason string `json:"reason"`
	TimeStamp int64 `json:"timestamp"`
}

func PublishEvent(event AuthEvent) error{
	if writer!=nil {
		payload, _ :=json.Marshal(event)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := writer.WriteMessages(ctx,kafka.Message{Value: payload})
		if err != nil {
			log.Printf("[KAFKA ERROR] Failed to write message: %v", err)
		} else {
			log.Printf("[KAFKA] Successfully wrote message for Gate %d", event.Gate)
		}
		return err
	}
	log.Printf("[KAFKA ERROR] Writer is nil")
	return nil
}

func CloseProducer() error{
	if writer!=nil{
		return writer.Close()
	}
	return nil
}
