FROM golang:1.20.4 AS BUILD
ENV GO111MODULE=auto
ARG SERVICE

WORKDIR /app

RUN echo ${SERVICE}

COPY src/${SERVICE}/ ./
COPY src/shared/ ./shared/
COPY config/config.yaml ./config.yaml

RUN go mod init main
RUN go mod tidy
#Optional
RUN go mod download


RUN go build -o main .

ENV PORT 5000

EXPOSE $PORT

ENTRYPOINT ["./main"]

EXPOSE 5000
