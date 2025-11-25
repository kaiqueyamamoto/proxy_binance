# Proxy Binance

Servidor proxy em Go para fazer proxy de requisições para a API da Binance, resolvendo problemas de CORS e restrições geográficas.

## 🚀 Funcionalidades

- Proxy completo para a API da Binance
- Framework Gin para alta performance
- Swagger UI integrado - documentação interativa
- Suporte a CORS para requisições do frontend
- Health check endpoint
- Teste de conexão com a Binance
- Configurável via variáveis de ambiente
- Timeout configurável para requisições
- Logs detalhados em formato debug

## 📋 Pré-requisitos

- Go 1.21 ou superior
- Acesso à internet para conectar com a API da Binance

## 🔧 Instalação

1. Navegue até a pasta do projeto:
```bash
cd proxy_binance
```

2. Instale as dependências:
```bash
go mod download
```

## 🏃 Execução

### Desenvolvimento

```bash
go run main.go
```

### Produção

```bash
go build -o binance-proxy main.go
./binance-proxy
```

### Com Docker (opcional)

```bash
docker build -t binance-proxy .
docker run -p 8080:8080 binance-proxy
```

## ⚙️ Configuração

### Variáveis de Ambiente

- `PORT`: Porta do servidor (padrão: `8080`)
- `BINANCE_API_URL`: URL da API da Binance (padrão: `https://api.binance.com/api/v3`)

### Exemplo

```bash
export PORT=3000
export BINANCE_API_URL=https://api.binance.com/api/v3
go run main.go
```

## 📡 Endpoints

### Health Check
```
GET /health
```
Retorna o status do proxy.

**Resposta:**
```json
{
  "status": "ok",
  "service": "binance-proxy",
  "time": "2025-11-25T18:00:00Z",
  "binance_url": "https://api.binance.com/api/v3"
}
```

### Test Connection
```
GET /test
```
Testa a conexão com a API da Binance.

**Resposta:**
```json
{
  "status": "ok",
  "binance_url": "https://api.binance.com/api/v3",
  "http_status": 200,
  "message": "Conexão com Binance estabelecida com sucesso"
}
```

### Proxy para API da Binance
```
GET /ticker/24hr
GET /ticker/24hr?symbol=BTCUSDT
GET /klines?symbol=BTCUSDT&interval=1h&limit=100
POST /api/v3/order
```

Todas as rotas são repassadas para a API da Binance.

## 📚 Documentação Swagger/OpenAPI

O projeto inclui documentação Swagger completa integrada ao servidor Gin. A documentação está disponível diretamente no servidor.

### Acessar a Documentação Swagger UI

Após iniciar o servidor, acesse:

```
http://localhost:8080/swagger/index.html
```

A documentação Swagger UI está totalmente integrada e permite:
- Visualizar todos os endpoints disponíveis
- Testar endpoints diretamente na interface
- Ver exemplos de requisições e respostas
- Entender os parâmetros necessários

### Endpoint JSON da Documentação

O arquivo JSON da documentação está disponível em:

```
http://localhost:8080/swagger/doc.json
```

### Principais Endpoints Documentados

- **Proxy:**
  - `GET /health` - Health check
  - `GET /test` - Testar conexão com Binance

- **Market Data:**
  - `GET /ticker/24hr` - Estatísticas de 24 horas
  - `GET /ticker/price` - Preço atual
  - `GET /ticker/bookTicker` - Melhor preço de compra/venda
  - `GET /klines` - Dados de candlestick
  - `GET /depth` - Livro de ordens
  - `GET /trades` - Negociações recentes
  - `GET /avgPrice` - Preço médio
  - `GET /exchangeInfo` - Informações da exchange
  - `GET /ping` - Teste de conectividade
  - `GET /time` - Tempo do servidor

Consulte o arquivo `swagger.yaml` para a documentação completa de todos os endpoints e parâmetros.

## 🔗 Uso no Frontend

### Exemplo com fetch

```javascript
// Substituir a URL da Binance pela URL do proxy
const PROXY_URL = 'http://localhost:8080';

// Buscar ticker
const response = await fetch(`${PROXY_URL}/ticker/24hr?symbol=BTCUSDT`);
const data = await response.json();

// Buscar klines
const klines = await fetch(
  `${PROXY_URL}/klines?symbol=BTCUSDT&interval=1h&limit=100`
);
const klinesData = await klines.json();
```

### Exemplo com Next.js

No arquivo `src/app/api/crypto/route.ts`, você pode usar:

```typescript
const PROXY_URL = process.env.BINANCE_PROXY_URL || 'http://localhost:8080';
const response = await fetch(`${PROXY_URL}/ticker/24hr`);
```

## 🐳 Docker

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o binance-proxy main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/binance-proxy .

EXPOSE 8080
CMD ["./binance-proxy"]
```

### Build e Run

```bash
docker build -t binance-proxy .
docker run -p 8080:8080 -e PORT=8080 binance-proxy
```

## 📝 Estrutura do Projeto

```
proxy_binance/
├── main.go          # Código principal do proxy (Gin + Swagger)
├── go.mod           # Dependências do Go
├── go.sum           # Checksums das dependências
├── swagger.yaml     # Documentação Swagger/OpenAPI
├── SWAGGER.md       # Guia de uso do Swagger
├── README.md        # Este arquivo
└── .gitignore       # Arquivos ignorados pelo Git
```

## 🔒 Segurança

- O proxy não armazena nenhuma informação sensível
- Todas as requisições são repassadas diretamente para a Binance
- CORS configurado para permitir requisições de qualquer origem (ajuste conforme necessário)

## 🐛 Troubleshooting

### Erro de conexão com Binance

Verifique se:
1. Você tem acesso à internet
2. A URL da Binance está correta
3. Não há firewall bloqueando a conexão

### Erro de CORS no frontend

Certifique-se de que o proxy está rodando e a URL está correta no frontend.

### Porta já em uso

Altere a porta usando a variável de ambiente `PORT`:
```bash
PORT=3000 go run main.go
```

## 📄 Licença

Este projeto é parte do projeto principal e segue a mesma licença.

