package kafka

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	
	"shield/thread-intel/internal/config"
	"shield/thread-intel/internal/store"
)

type Consumer struct {
	consumer *kafka.Consumer
	store    *store.ThreatStore
	cfg      config.Config
}

func NewConsumer(cfg config.Config, ts *store.ThreatStore) (*Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBrokerUrl,
		"group.id":          cfg.KafkaGroupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: c,
		store:    ts,
		cfg:      cfg,
	}, nil
}

func (c *Consumer) Start() error {
	err := c.consumer.SubscribeTopics([]string{c.cfg.KafkaTopic}, nil)
	if err != nil {
		return err
	}

	log.Printf("Listening on topic: %s", c.cfg.KafkaTopic)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	for run {
		select {
		case sig := <-sigchan:
			log.Printf("Caught signal %v: terminating", sig)
			run = false
		default:
			ev := c.consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				log.Printf("Received message on %s: %s", e.TopicPartition, string(e.Value))
				var event store.UrlEvent
				if err := json.Unmarshal(e.Value, &event); err != nil {
					log.Printf("Failed to unmarshal event: %v", err)
					continue
				}

				if err := c.store.InsertThreat(event); err != nil {
					log.Printf("Failed to insert into DB: %v", err)
				} else {
					log.Println("Successfully saved threat to database!")
				}
			case kafka.Error:
				log.Printf("Kafka error: %v", e)
			}
		}
	}

	return nil
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
