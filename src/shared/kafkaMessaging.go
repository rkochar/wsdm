package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

const KAFKA_SERVICE = "kafka-service:9092"

//const KAFKA_SERVICE = "localhost:9092"

func SetUpKafkaListener(services []string, inLockMaster bool, action func(*SagaMessage) (*SagaMessage, string)) {
	// Set up Kafka connection configuration
	brokers := []string{KAFKA_SERVICE}

	// Set up Kafka reader configuration
	config := kafka.ReaderConfig{
		Brokers:         brokers,
		GroupID:         "saga-orchestrator-group",
		MinBytes:        10e3,
		MaxBytes:        10e6,
		MaxWait:         1 * time.Second,
		ReadLagInterval: -1,
	}

	readerMap := make(map[string]*kafka.Reader)
	senderMap := make(map[string]*kafka.Conn)

	var sendName string
	var receiveName string

	if inLockMaster {
		sendName = "-syn"
		receiveName = "-ack"
	} else {
		sendName = "-ack"
		receiveName = "-syn"
	}

	for _, serviceName := range services {
		sendTopic := serviceName + sendName
		senderMap[sendTopic] = CreateTopicSender(sendTopic)
		defer senderMap[sendTopic].Close()

		receiveTopic := serviceName + receiveName
		readerMap[receiveTopic] = CreateTopicReader(receiveTopic, config)
		defer readerMap[receiveTopic].Close()
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

					log.Printf("Received message for topic %s: %s\n", topic, string(m.Value))

					parseErr, message := ParseSagaMessage(string(m.Value))
					if parseErr != nil {
						log.Printf("Error parsing message: %s\n", parseErr)
						// log.Printf("Error parsing message: %s\n", parseErr)
						continue
					}

					returnMessage, senderName := action(message)

					if returnMessage == nil || senderName == "" {
						continue
					}

					log.Printf("Sending message to topic %s: %s_%d\n", senderName, returnMessage.Name, returnMessage.SagaID)

					sendErr := SendSagaMessage(returnMessage, senderMap[senderName])
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

func CreateTopicSender(topic string) *kafka.Conn {
	conn, err := kafka.DialLeader(context.Background(), "tcp", KAFKA_SERVICE, topic, 0)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	return conn
}

func CreateTopicReader(topicName string, config kafka.ReaderConfig) *kafka.Reader {
	config.Topic = topicName
	reader := kafka.NewReader(config)
	return reader
}

func SendSagaMessage(message *SagaMessage, conn *kafka.Conn) error {
	jsonByteArray, marshalError := json.Marshal(message.Order)
	if marshalError != nil {
		return marshalError
	}

	messageBuffer := bytes.Buffer{}
	messageBuffer.WriteString(message.Name)
	messageBuffer.WriteRune('_')
	messageBuffer.WriteString(strconv.FormatInt(message.SagaID, 10))
	messageBuffer.WriteRune('_')
	messageBuffer.Write(jsonByteArray)

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, writeErr := conn.Write(messageBuffer.Bytes())
	if writeErr != nil {
		return writeErr
	}
	return nil
}
