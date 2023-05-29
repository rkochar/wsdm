FROM golang:1.20.4 AS BUILD

WORKDIR /app

COPY go.mod go.sum app.go status.go utils.go ./

RUN go mod download && \
    go build -o main

FROM golang:1.20.4

WORKDIR /app

COPY --from=build /app/main .

ENV PORT 5000

EXPOSE $PORT

ENTRYPOINT ["./main"]
