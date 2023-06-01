module main

go 1.18

replace WDM-G1/shared => ../shared

require (
<<<<<<< HEAD:src/go.mod
	github.com/go-sql-driver/mysql v1.7.1
=======
	WDM-G1/shared v0.0.0-00010101000000-000000000000
>>>>>>> main:payment/go.mod
	github.com/gorilla/mux v1.8.0
	github.com/labstack/echo/v4 v4.10.2
	github.com/segmentio/kafka-go v0.4.40
	go.mongodb.org/mongo-driver v1.11.6
)

require (
	github.com/golang/snappy v0.0.1 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/labstack/gommon v0.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	golang.org/x/crypto v0.6.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
)
