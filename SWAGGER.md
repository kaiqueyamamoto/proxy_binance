# Documentação Swagger - Binance Proxy API

Este documento explica como visualizar e usar a documentação Swagger/OpenAPI do Binance Proxy.

## 📋 Sobre o Swagger

O arquivo `swagger.yaml` contém a documentação completa da API do proxy, incluindo:

- **Endpoints do Proxy**: `/health`, `/test`
- **Principais Endpoints da Binance**: 
  - Market Data (ticker, klines, depth, trades, etc.)
  - Informações da Exchange
  - Dados de tempo e conectividade

## 🚀 Como Visualizar

### Método 1: Swagger Editor Online (Mais Fácil)

1. Acesse https://editor.swagger.io/
2. Clique em "File" → "Import file" ou cole o conteúdo do `swagger.yaml`
3. A documentação será renderizada automaticamente
4. Você pode testar os endpoints diretamente na interface

**Vantagens:**
- Não requer instalação
- Interface interativa
- Permite testar endpoints

### Método 2: Swagger UI com Docker

```bash
# Executar Swagger UI em um container Docker
docker run -p 8081:8080 \
  -e SWAGGER_JSON=/swagger.yaml \
  -v $(pwd)/swagger.yaml:/swagger.yaml \
  swaggerapi/swagger-ui
```

Depois acesse: http://localhost:8081

**Vantagens:**
- Interface completa do Swagger UI
- Permite testar endpoints
- Não polui seu ambiente local

### Método 3: Swagger UI com npm/npx

```bash
# Instalar globalmente (opcional)
npm install -g swagger-ui-serve

# Ou usar diretamente com npx
npx swagger-ui-serve swagger.yaml
```

**Vantagens:**
- Rápido e simples
- Não requer Docker

### Método 4: Redoc (Interface Alternativa)

```bash
# Com npx
npx @redocly/cli preview-docs swagger.yaml

# Ou instalar globalmente
npm install -g @redocly/cli
redocly preview-docs swagger.yaml
```

**Vantagens:**
- Interface mais limpa e moderna
- Melhor para documentação

### Método 5: VS Code Extension

1. Instale a extensão "OpenAPI (Swagger) Editor" no VS Code
2. Abra o arquivo `swagger.yaml`
3. Use o comando "OpenAPI: Preview" (Ctrl+Shift+P)

**Vantagens:**
- Integrado ao editor
- Validação em tempo real

## 🧪 Testando Endpoints

### No Swagger UI

1. Abra a documentação em qualquer um dos métodos acima
2. Expanda o endpoint desejado
3. Clique em "Try it out"
4. Preencha os parâmetros (se necessário)
5. Clique em "Execute"
6. Veja a resposta na interface

### Com cURL

Exemplos baseados na documentação:

```bash
# Health check
curl http://localhost:8080/health

# Ticker 24h para BTCUSDT
curl "http://localhost:8080/ticker/24hr?symbol=BTCUSDT"

# Klines (candlestick)
curl "http://localhost:8080/klines?symbol=BTCUSDT&interval=1h&limit=100"

# Order book
curl "http://localhost:8080/depth?symbol=BTCUSDT&limit=20"

# Preço atual
curl "http://localhost:8080/ticker/price?symbol=BTCUSDT"
```

### Com JavaScript/Fetch

```javascript
// Exemplo: Buscar ticker 24h
const response = await fetch('http://localhost:8080/ticker/24hr?symbol=BTCUSDT');
const data = await response.json();
console.log(data);

// Exemplo: Buscar klines
const klines = await fetch(
  'http://localhost:8080/klines?symbol=BTCUSDT&interval=1h&limit=100'
);
const klinesData = await klines.json();
console.log(klinesData);
```

## 📖 Estrutura da Documentação

### Tags

- **Proxy**: Endpoints do próprio proxy (`/health`, `/test`)
- **Market Data**: Dados de mercado públicos (não requerem autenticação)
- **Account**: Dados da conta (requerem autenticação - não implementado ainda)

### Schemas

A documentação define vários schemas reutilizáveis:

- `Error`: Formato padrão de erro
- `Ticker24hr`: Dados de ticker de 24 horas
- `PriceTicker`: Preço atual de um símbolo
- `BookTicker`: Melhor preço de compra/venda
- `AvgPrice`: Preço médio
- `OrderBook`: Livro de ordens
- `Trade`: Dados de uma negociação

## 🔧 Personalização

Para adicionar novos endpoints ao Swagger:

1. Abra o arquivo `swagger.yaml`
2. Adicione o novo endpoint na seção `paths`
3. Defina os parâmetros, respostas e schemas necessários
4. Atualize a visualização

### Exemplo: Adicionar um novo endpoint

```yaml
  /novo-endpoint:
    get:
      tags:
        - Market Data
      summary: Novo Endpoint
      description: Descrição do novo endpoint
      operationId: novoEndpoint
      parameters:
        - name: symbol
          in: query
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Resposta de sucesso
          content:
            application/json:
              schema:
                type: object
```

## 📚 Recursos Adicionais

- [Documentação Oficial do OpenAPI](https://swagger.io/specification/)
- [Swagger Editor](https://editor.swagger.io/)
- [Binance API Documentation](https://binance-docs.github.io/apidocs/spot/en/)
- [Swagger UI GitHub](https://github.com/swagger-api/swagger-ui)

## 🐛 Troubleshooting

### Erro ao abrir no Swagger Editor

- Verifique se o YAML está bem formatado
- Use um validador YAML online
- Certifique-se de que não há caracteres especiais inválidos

### Endpoints não funcionam no Swagger UI

- Certifique-se de que o proxy está rodando
- Verifique se a URL base está correta no `swagger.yaml`
- Alguns endpoints podem requerer autenticação (não implementado)

### Docker não inicia

- Verifique se a porta 8081 está disponível
- Tente usar outra porta: `-p 8082:8080`
- Verifique se o arquivo `swagger.yaml` está no diretório correto

## 💡 Dicas

1. **Use o Swagger Editor** para validação em tempo real enquanto edita
2. **Mantenha a documentação atualizada** quando adicionar novos endpoints
3. **Teste os endpoints** diretamente no Swagger UI antes de usar no código
4. **Compartilhe a documentação** com sua equipe para facilitar a integração

