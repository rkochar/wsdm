FROM golang:1.20.4 AS BUILD
ENV GO111MODULE=auto
ARG SERVICE

WORKDIR /app

RUN echo ${SERVICE}

COPY ${SERVICE}/go.mod ${SERVICE}/go.sum ${SERVICE}/app.go ${SERVICE}/utils.go ./service/
COPY /shared/ ./shared/


WORKDIR /app/shared
RUN go mod download

WORKDIR /app/service
RUN go mod download




WORKDIR /app/service
#RUN ls --recursive / && sleep 15
RUN go build -o main
#

ENV PORT 5000

EXPOSE $PORT

ENTRYPOINT ["./main"]

EXPOSE 5000
