package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const (
	defaultPort        = "8080"
	binanceAPIBaseURL  = "https://api.binance.com/api/v3"
	readTimeout        = 30 * time.Second
	writeTimeout       = 30 * time.Second
)

type ProxyServer struct {
	binanceURL string
	client     *http.Client
}

func NewProxyServer() *ProxyServer {
	return &ProxyServer{
		binanceURL: binanceAPIBaseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyRequest faz o proxy da requisição para a API da Binance
func (p *ProxyServer) ProxyRequest(w http.ResponseWriter, r *http.Request) {
	// Obter o path da requisição (ex: /ticker/24hr, /klines, etc.)
	path := r.URL.Path
	
	// Remover o prefixo /api se existir
	if strings.HasPrefix(path, "/api") {
		path = strings.TrimPrefix(path, "/api")
	}
	
	// Construir a URL completa da Binance
	targetURL := fmt.Sprintf("%s%s", p.binanceURL, path)
	
	// Adicionar query parameters se existirem
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	log.Printf("Proxying request: %s %s -> %s", r.Method, r.URL.Path, targetURL)

	// Criar a requisição para a Binance
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao criar requisição: %v", err), http.StatusInternalServerError)
		return
	}

	// Copiar headers importantes (exceto Host)
	for key, values := range r.Header {
		if strings.ToLower(key) != "host" {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	// Fazer a requisição para a Binance
	resp, err := p.client.Do(req)
	if err != nil {
		log.Printf("Erro ao fazer requisição para Binance: %v", err)
		http.Error(w, fmt.Sprintf("Erro ao conectar com Binance: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Ler o corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler resposta da Binance: %v", err)
		http.Error(w, "Erro ao ler resposta", http.StatusInternalServerError)
		return
	}

	// Copiar status code e headers da resposta
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	
	// Copiar outros headers importantes
	for key, values := range resp.Header {
		if strings.ToLower(key) != "content-encoding" {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// HealthCheck endpoint para verificar se o proxy está funcionando
func (p *ProxyServer) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "ok",
		"service": "binance-proxy",
		"time":    time.Now().Format(time.RFC3339),
		"binance_url": p.binanceURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TestConnection testa a conexão com a Binance
func (p *ProxyServer) TestConnection(w http.ResponseWriter, r *http.Request) {
	testURL := fmt.Sprintf("%s/ping", p.binanceURL)
	
	resp, err := p.client.Get(testURL)
	if err != nil {
		response := map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Erro ao conectar com Binance: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer resp.Body.Close()

	response := map[string]interface{}{
		"status":      "ok",
		"binance_url": p.binanceURL,
		"http_status": resp.StatusCode,
		"message":     "Conexão com Binance estabelecida com sucesso",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Obter porta do ambiente ou usar padrão
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Permitir override da URL da Binance via variável de ambiente
	binanceURL := os.Getenv("BINANCE_API_URL")
	if binanceURL == "" {
		binanceURL = binanceAPIBaseURL
	}

	proxy := &ProxyServer{
		binanceURL: binanceURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Configurar router
	router := mux.NewRouter()

	// Rotas
	router.HandleFunc("/health", proxy.HealthCheck).Methods("GET")
	router.HandleFunc("/test", proxy.TestConnection).Methods("GET")
	
	// Proxy para todas as rotas da API da Binance
	router.PathPrefix("/").HandlerFunc(proxy.ProxyRequest)

	// Configurar CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		AllowCredentials: false,
	})

	handler := c.Handler(router)

	// Configurar servidor HTTP
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	log.Printf("🚀 Proxy Binance iniciado na porta %s", port)
	log.Printf("📡 URL da Binance: %s", binanceURL)
	log.Printf("🌐 Endpoints disponíveis:")
	log.Printf("   - GET  /health - Health check")
	log.Printf("   - GET  /test - Testar conexão com Binance")
	log.Printf("   - GET  /* - Proxy para API da Binance")
	log.Printf("   - POST /* - Proxy para API da Binance")
	
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

