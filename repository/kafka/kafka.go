package kafka

import (
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type Producer struct {
	Prod  sarama.AsyncProducer
	Topic string
}

type TransactionEvent struct {
	UserID    int64     `json:"user_id"`
	Entry     string    `json:"entry"`
	Amount    int64     `json:"amount"`
	Balance   int64     `json:"balance"`
	Timestamp time.Time `json:"timestamp"`
}

func ConnectKafka(brokersUrl ...string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Idempotent = true

	producer, err := sarama.NewAsyncProducer(brokersUrl, config)

	if err != nil {
		return nil, err
	}
	prod := &Producer{Prod: producer, Topic: "transaction"}

	go func() {
		for {
			select {
			case success := <-prod.Prod.Successes():
				log.Printf("Kafka message sent to partition %d at offset %d\n", success.Partition, success.Offset)
			case err := <-prod.Prod.Errors():
				log.Printf("Kafka producer error: %v\n", err.Err)
			}
		}
	}()
	return prod, nil
}

func (kp *Producer) PublishTransaction(event TransactionEvent) {
	bytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal transaction event: %v\n", err)
		return
	}

	kp.Prod.Input() <- &sarama.ProducerMessage{
		Topic: kp.Topic,
		Value: sarama.ByteEncoder(bytes),
	}
}

func (kp *Producer) Close() {
	kp.Prod.Close()
}
