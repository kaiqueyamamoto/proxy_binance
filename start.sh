#!/bin/bash

# Script para iniciar o proxy Binance

# Cores para output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Iniciando Proxy Binance...${NC}"

# Verificar se Go está instalado
if ! command -v go &> /dev/null; then
    echo "❌ Go não está instalado. Por favor, instale o Go primeiro."
    exit 1
fi

# Verificar se as dependências estão instaladas
if [ ! -f "go.sum" ]; then
    echo "📦 Instalando dependências..."
    go mod download
fi

# Definir porta padrão se não estiver definida
export PORT=${PORT:-8080}
export BINANCE_API_URL=${BINANCE_API_URL:-https://api.binance.com/api/v3}

echo -e "${GREEN}✅ Configuração:${NC}"
echo "   Porta: $PORT"
echo "   URL Binance: $BINANCE_API_URL"
echo ""

# Executar o proxy
go run main.go

