package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

const KAFKA_SERVICE = "kafka-service.kafka:9092"

func SetUpKafkaListener(services []string, inLockMaster bool, action func(*SagaMessage) (*SagaMessage, string)) {

	partition := 1

	readerMap := make(map[string]*kafka.Conn)
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
		senderMap[sendTopic] = CreateConnection(sendTopic, partition)
		defer senderMap[sendTopic].Close()

		receiveTopic := serviceName + receiveName
		readerMap[receiveTopic] = CreateConnection(receiveTopic, partition)
		defer readerMap[receiveTopic].Close()
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for topic, reader := range readerMap {
		go func(reader *kafka.Conn) {
			for {
				select {
				case <-signals:
					log.Printf("Received interrupt signal for topic %s. Shutting down...\n", topic)
					return
				default:

					reader.SetReadDeadline(time.Now().Add(10 * time.Second))
					batch := reader.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max

					b := make([]byte, 10e3) // 10KB max per message
					for {
						n, err := batch.Read(b)

						if err != nil {
							if strings.Contains(err.Error(), "context canceled") {
								log.Printf("Consumer context canceled for topic %s. Shutting down...\n", topic)
								return
							}
							log.Printf("Error reading message for topic %s: %v\n", topic, err)
							continue
						}

						parseErr, message := ParseSagaMessage(string(b[:n]))
						if parseErr != nil {
							fmt.Printf("Error parsing message: %s\n", parseErr)
							continue
						}

						returnMessage, senderName := action(message)

						if returnMessage == nil {
							continue
						}

						sendErr := SendSagaMessage(returnMessage, senderMap[senderName])
						if sendErr != nil {
							log.Printf("Error sending message: %s\n", sendErr)
						}
					}

					if err := batch.Close(); err != nil {
						log.Fatal("failed to close batch:", err)
					}
				}
			}
		}(reader)
	}

	// Wait for termination signal
	<-signals
	log.Println("Received interrupt signal. Shutting down...")
}

func CreateConnection(topic string, partition int) *kafka.Conn {
	log.Print("CreateConnection")
	log.Print("topic:", topic, "\npartition:", partition, "\nservice:", KAFKA_SERVICE)
	conn, err := kafka.DialLeader(context.Background(), "tcp", KAFKA_SERVICE, topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	return conn
}

func SendSagaMessage(message *SagaMessage, conn *kafka.Conn) error {
	fmt.Println("message.Order:", message.Order)
	jsonByteArray, marshalError := json.Marshal(message.Order)
	if marshalError != nil {
		return marshalError
	}

	messageBuffer := bytes.Buffer{}
	messageBuffer.WriteString("START_CHECKOUT-SAGA_")
	messageBuffer.WriteString(strconv.FormatInt(message.SagaID, 10))
	messageBuffer.WriteString("_")
	messageBuffer.Write(jsonByteArray)

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, writeErr := conn.Write(messageBuffer.Bytes())
	if writeErr != nil {
		return writeErr
	}
	return nil
}
