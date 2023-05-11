FROM golang:1.20.4 AS BUILD

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY app.go .
COPY utils.go .

RUN go build -o main

FROM golang:1.20.4

WORKDIR /app

COPY --from=build /app/main .

ENV PORT 5000

EXPOSE $PORT

ENTRYPOINT ["./main"]