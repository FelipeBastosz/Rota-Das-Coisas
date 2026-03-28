# Estágio 1: Compilação
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copia o go.mod da raiz do projeto
COPY go.mod ./

RUN go mod download

#Copia o projeto para dentro do container
# (Isso funciona porque o ponto final do comando lá em cima é a raiz)
COPY . .

# Compila o binário apontando o caminho relativo à raiz
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./Server/server.go

# Estágio 2: Execução
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/server .

# Expõe as portas
EXPOSE 8080/tcp
EXPOSE 5000/udp

# Comando para rodar o servidor
CMD ["./server"]