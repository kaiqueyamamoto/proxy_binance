FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o binance-proxy main.go

# Stage final
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiar o binário do builder
COPY --from=builder /app/binance-proxy .

# Expor porta
EXPOSE 8080

# Variáveis de ambiente padrão
ENV PORT=8080
ENV BINANCE_API_URL=https://api.binance.com/api/v3

# Comando para executar
CMD ["./binance-proxy"]

