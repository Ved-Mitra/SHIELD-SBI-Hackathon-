package kafka

import (
	"auth-intel/internal/config"
	"auth-intel/internal/store"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	store *store.AuthStore
	cfg config.Config
}

func NewConsumer(cfg config.Config, as* store.AuthStore) (*Consumer, error){
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaBrokerUrl},
		GroupID: cfg.KafkaGroupID,
		Topic: cfg.KafkaTopic,
	})

	return &Consumer{
		reader: reader,
		store: as,
		cfg: cfg,
	}, nil
}

func (c *Consumer) Start() error{
	log.Printf("Listening To topic: %s",c.cfg.KafkaTopic)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	sigchan := make(chan os.Signal,1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func(){
		<-sigchan
		log.Printf("Caught signal: terminating")
		cancel()
	}()

	for{
		m,err := c.reader.ReadMessage(ctx)
		if err!=nil{
			if ctx.Err()!=nil{
				break
			}
			log.Printf("Kafka read error: %v", err)
		}

		log.Printf("Received message on %s: %s", m.Topic,string(m.Value))

		var event store.AuthEvent
		if err:=json.Unmarshal(m.Value,&event); err!=nil{
			log.Printf("Failed to unmarshall event: %v",err)
			continue
		}

		if err:= c.store.InsertAuth(event); err!=nil{
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