module github.com/rkochar/wsdm

go 1.20

replace WDM-G1/shared => ../shared

require (
	WDM-G1/shared v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.7.1
)

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/segmentio/kafka-go v0.4.40 // indirect
	go.mongodb.org/mongo-driver v1.11.6 // indirect
)
