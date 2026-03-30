FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o actuator ./Actuators/actuator.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/actuator .
CMD ["./actuator", "atuador_padrao"]