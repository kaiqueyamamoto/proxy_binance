# Integração do Proxy Binance com o Frontend

Este documento explica como integrar o proxy Binance com o frontend Next.js.

## 🚀 Passo 1: Iniciar o Proxy

### Desenvolvimento Local

```bash
cd proxy_binance
go run main.go
```

Ou use os scripts:

**Windows:**
```bash
start.bat
```

**Linux/Mac:**
```bash
chmod +x start.sh
./start.sh
```

O proxy estará rodando em `http://localhost:8080`

## 🔧 Passo 2: Configurar Variáveis de Ambiente

No arquivo `.env.local` do projeto Next.js, adicione:

```env
# URL do proxy Binance
BINANCE_PROXY_URL=http://localhost:8080

# Ou em produção, use a URL do seu servidor
# BINANCE_PROXY_URL=https://seu-dominio.com
```

## 📝 Passo 3: Atualizar o Código do Frontend

### Opção 1: Atualizar `src/app/api/crypto/route.ts`

```typescript
export async function GET(request: NextRequest) {
  try {
    const searchParams = request.nextUrl.searchParams;
    const limit = parseInt(searchParams.get("limit") || "100");
    const quote = searchParams.get("quote") || "USDT";

    // Usar o proxy em vez da API direta da Binance
    const PROXY_URL = process.env.BINANCE_PROXY_URL || "http://localhost:8080";
    const BINANCE_API_URL = `${PROXY_URL}`;

    // Buscar todos os tickers diretamente através do proxy
    const tickerResponse = await fetch(`${BINANCE_API_URL}/ticker/24hr`, {
      cache: "no-store",
    });

    if (!tickerResponse.ok) {
      throw new Error(`Binance API error: ${tickerResponse.statusText}`);
    }

    const allTickers = await tickerResponse.json();
    
    // ... resto do código
  } catch (error) {
    // ... tratamento de erros
  }
}
```

### Opção 2: Atualizar `src/app/crypto/page.tsx`

Substitua todas as URLs da Binance:

```typescript
// No início do componente, adicione:
const PROXY_URL = process.env.NEXT_PUBLIC_BINANCE_PROXY_URL || "http://localhost:8080";

// Substitua:
// `https://api.binance.com/api/v3/ticker/24hr?symbol=${symbol}`
// Por:
`${PROXY_URL}/ticker/24hr?symbol=${symbol}`

// Substitua:
// `https://api.binance.com/api/v3/klines?symbol=${symbol}&interval=${interval}&limit=${limit}`
// Por:
`${PROXY_URL}/klines?symbol=${symbol}&interval=${interval}&limit=${limit}`
```

## 🌐 Passo 4: Deploy do Proxy

### Opção A: Deploy em Servidor Separado

1. Build do proxy:
```bash
cd proxy_binance
go build -o binance-proxy main.go
```

2. Execute em um servidor com acesso à internet:
```bash
./binance-proxy
```

3. Configure a URL no `.env` do Next.js:
```env
BINANCE_PROXY_URL=https://proxy.seu-dominio.com
```

### Opção B: Deploy com Docker

1. Build da imagem:
```bash
cd proxy_binance
docker build -t binance-proxy .
```

2. Execute o container:
```bash
docker run -d -p 8080:8080 \
  -e PORT=8080 \
  -e BINANCE_API_URL=https://api.binance.com/api/v3 \
  --name binance-proxy \
  binance-proxy
```

### Opção C: Deploy na Vercel (usando Serverless Function)

Crie um arquivo `api/proxy/[...path]/route.ts`:

```typescript
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  const path = params.path.join('/');
  const searchParams = request.nextUrl.searchParams.toString();
  const queryString = searchParams ? `?${searchParams}` : '';
  
  const binanceUrl = `https://api.binance.com/api/v3/${path}${queryString}`;
  
  try {
    const response = await fetch(binanceUrl, {
      cache: 'no-store',
    });
    
    const data = await response.json();
    
    return NextResponse.json(data, {
      status: response.status,
      headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type',
      },
    });
  } catch (error) {
    return NextResponse.json(
      { error: 'Erro ao fazer proxy para Binance' },
      { status: 500 }
    );
  }
}
```

## ✅ Testando a Integração

1. Verifique se o proxy está rodando:
```bash
curl http://localhost:8080/health
```

2. Teste uma requisição:
```bash
curl http://localhost:8080/ticker/24hr?symbol=BTCUSDT
```

3. Teste no frontend:
```javascript
const response = await fetch('http://localhost:8080/ticker/24hr?symbol=BTCUSDT');
const data = await response.json();
console.log(data);
```

## 🔒 Segurança em Produção

1. **Restringir CORS**: Ajuste as origens permitidas no `main.go`:
```go
c := cors.New(cors.Options{
    AllowedOrigins: []string{"https://seu-dominio.com"},
    // ...
})
```

2. **Rate Limiting**: Considere adicionar rate limiting para evitar abuso.

3. **Autenticação**: Se necessário, adicione autenticação para proteger o proxy.

## 🐛 Troubleshooting

### Erro: "Connection refused"
- Verifique se o proxy está rodando
- Verifique se a porta está correta

### Erro: "CORS policy"
- Verifique se o proxy está configurado para permitir CORS
- Verifique se a URL do proxy está correta

### Erro: "Timeout"
- Aumente o timeout no `main.go`
- Verifique sua conexão com a internet

