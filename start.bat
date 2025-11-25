@echo off
REM Script para iniciar o proxy Binance no Windows

echo 🚀 Iniciando Proxy Binance...

REM Verificar se Go está instalado
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Go não está instalado. Por favor, instale o Go primeiro.
    exit /b 1
)

REM Verificar se as dependências estão instaladas
if not exist "go.sum" (
    echo 📦 Instalando dependências...
    go mod download
)

REM Definir porta padrão se não estiver definida
if "%PORT%"=="" set PORT=8080
if "%BINANCE_API_URL%"=="" set BINANCE_API_URL=https://api.binance.com/api/v3

echo ✅ Configuração:
echo    Porta: %PORT%
echo    URL Binance: %BINANCE_API_URL%
echo.

REM Executar o proxy
go run main.go

