package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
	
	"shield/thread-intel/internal/config"
	"shield/thread-intel/internal/store"
)

type Consumer struct {
	reader *kafka.Reader
	store  *store.ThreatStore
	cfg    config.Config
}

func NewConsumer(cfg config.Config, ts *store.ThreatStore) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaBrokerUrl},
		GroupID: cfg.KafkaGroupID,
		Topic:   cfg.KafkaTopic,
	})

	return &Consumer{
		reader: reader,
		store:  ts,
		cfg:    cfg,
	}, nil
}

func (c *Consumer) Start() error {
	log.Printf("Listening on topic: %s", c.cfg.KafkaTopic)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigchan
		log.Println("Caught signal: terminating")
		cancel()
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				// Context cancelled, clean exit
				break
			}
			log.Printf("Kafka read error: %v", err)
			continue
		}

		log.Printf("Received message on %s: %s", m.Topic, string(m.Value))
		
		var event store.UrlEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}

		if err := c.store.InsertThreat(event); err != nil {
			log.Printf("Failed to insert into DB: %v", err)
		} else {
			log.Println("Successfully saved threat to database!")
		}
	}

	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
