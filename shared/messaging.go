package shared

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

func SetUpKafkaListener(services []string, action func(*SagaMessage) (*SagaMessage, string)) {
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

	senderMap := make(map[string]*kafka.Conn)
	readerMap := make(map[string]*kafka.Reader)

	for _, serviceName := range services {
		topicSyn := serviceName + "-syn"
		senderMap[topicSyn] = createTopicSender(topicSyn)
		defer senderMap[topicSyn].Close()

		topicAck := serviceName + "-ack"
		readerMap[topicAck] = createTopicReader(topicAck, config)
		defer readerMap[topicAck].Close()
	}

	// Create a context to control the consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful termination
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for _, reader := range readerMap {
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

					parseErr, message := parseSagaMessage(string(m.Value))
					if parseErr != nil {
						log.Printf("Error parsing message: %s\n", parseErr)
						continue
					}

					returnMessage, senderName := action(message)

					sendErr := sendSagaMessage(returnMessage, senderMap[senderName])
					if sendErr != nil {
						log.Printf("Error sending message: %s\n", sendErr)
					}
				}
			}
		}(reader)
	}

	// Wait for termination signal
	<-signals
	log.Println("Received interrupt signal. Shutting down...")
}

func createTopicSender(topic string) *kafka.Conn {
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, 0)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	return conn
}

func createTopicReader(topicName string, config kafka.ReaderConfig) *kafka.Reader {
	config.Topic = topicName
	reader := kafka.NewReader(config)
	return reader
}
