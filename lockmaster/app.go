package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type Action struct {
	nextMessage string
	topic       string
}

// Maps incoming message to outgoing message
var successfulActionMap = map[string]Action{
	// Normal checkout
	"START-CHECKOUT-SAGA": {"START-SUBTRACT-STOCK", "order-syn"},
	"END-SUBTRACT-STOCK":  {"START-MAKE-PAYMENT", "payment-syn"},
	"END-MAKE-PAYMENT":    {"START-UPDATE-ORDER", "order-syn"},
	"END-UPDATE-ORDER":    {"END-CHECKOUT-SAGA", ""},
	// Rollback checkout
	"END-CANCEL-PAYMENT": {"START-READD-STOCK", "stock-syn"},
	"END-READD-STOCK":    {"END-CHECKOUT-SAGA", ""},
}

// Maps message before ABORT to outgoing message
var failActionMap = map[string]Action{
	// Stock Fails
	"START-SUBTRACT-STOCK": {"END-CHECKOUT-SAGA", ""},
	// Payment Fails
	"START-MAKE-PAYMENT": {"START-READD-STOCK", "stock-syn"},
	// Order Fails
	"START-UPDATE-ORDER": {"START-CANCEL-PAYMENT", "payment-syn"},
}

func main() {
	// Set up Kafka connection configuration
	brokers := []string{"localhost:9092"}

	// Set up Kafka reader configuration
	config := kafka.ReaderConfig{
		Brokers:         brokers,
		GroupID:         "saga-orchestrator-group",
		MinBytes:        10e3,
		MaxBytes:        10e6,
		MaxWait:         1 * time.Second,
		ReadLagInterval: -1,
	}
	// Create the order-ack Reader
	orderReader := createTopicReader("order-ack", config)
	defer orderReader.Close()
	// Create the stock-ack Reader
	stockReader := createTopicReader("stock-ack", config)
	defer stockReader.Close()
	// Create the payment-ack Reader
	paymentReader := createTopicReader("payment-ack", config)
	defer paymentReader.Close()

	// Create a context to control the consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful termination
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Setup the logic for the order-ack listener
	setupOrderTopicListener(orderReader, ctx, signals)
	// Setup the logic for the stock-ack listener
	setupStockTopicListener(stockReader, ctx, signals)
	// Setup the logic for the payment-ack listener
	setupPaymentTopicListener(paymentReader, ctx, signals)

	// Wait for termination signal
	<-signals
	log.Println("Received interrupt signal. Shutting down...")
}

func createTopicReader(topicName string, config kafka.ReaderConfig) *kafka.Reader {
	config.Topic = topicName
	reader := kafka.NewReader(config)
	return reader
}

func setupOrderTopicListener(reader *kafka.Reader, ctx context.Context, signals <-chan os.Signal) {
	go func(reader *kafka.Reader) {
		topic := reader.Config().Topic
		for {
			select {
			case <-signals:
				log.Printf("Received interrupt signal for topic %s. Shutting down...\n", topic)
				return
			default:
				m, err := reader.ReadMessage(ctx)
				if err != nil {
					if strings.Contains(err.Error(), "context canceled") {
						log.Printf("Consumer context canceled for topic %s. Shutting down...\n", topic)
						return
					}
					log.Printf("Error reading message for topic %s: %v\n", topic, err)
					continue
				}

				fmt.Printf("Received message for topic %s: Partition=%d, Offset=%d, Key=%s, Value=%s\n",
					topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

			}
		}
	}(reader)
}

func setupStockTopicListener(reader *kafka.Reader, ctx context.Context, signals <-chan os.Signal) {
	go func(reader *kafka.Reader) {
		topic := reader.Config().Topic
		for {
			select {
			case <-signals:
				log.Printf("Received interrupt signal for topic %s. Shutting down...\n", topic)
				return
			default:
				m, err := reader.ReadMessage(ctx)
				if err != nil {
					if strings.Contains(err.Error(), "context canceled") {
						log.Printf("Consumer context canceled for topic %s. Shutting down...\n", topic)
						return
					}
					log.Printf("Error reading message for topic %s: %v\n", topic, err)
					continue
				}

				fmt.Printf("Received message for topic %s: Partition=%d, Offset=%d, Key=%s, Value=%s\n",
					topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
			}
		}
	}(reader)
}

func setupPaymentTopicListener(reader *kafka.Reader, ctx context.Context, signals <-chan os.Signal) {
	go func(reader *kafka.Reader) {
		topic := reader.Config().Topic
		for {
			select {
			case <-signals:
				log.Printf("Received interrupt signal for topic %s. Shutting down...\n", topic)
				return
			default:
				m, err := reader.ReadMessage(ctx)
				if err != nil {
					if strings.Contains(err.Error(), "context canceled") {
						log.Printf("Consumer context canceled for topic %s. Shutting down...\n", topic)
						return
					}
					log.Printf("Error reading message for topic %s: %v\n", topic, err)
					continue
				}

				fmt.Printf("Received message for topic %s: Partition=%d, Offset=%d, Key=%s, Value=%s\n",
					topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
			}
		}
	}(reader)
}
