FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o sensor ./Sensors/sensor.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/sensor .
# Deixamos um nome padrão, mas o Compose vai sobrescrever isso
CMD ["./sensor", "sensor_padrao"]