package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gopkg.in/yaml.v3"
)

const (
	defaultPort       = "8080"
	binanceAPIBaseURL = "https://api.binance.com/api/v3"
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second
)

// Função auxiliar para min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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
// @Summary Proxy para API da Binance
// @Description Repassa requisições para a API oficial da Binance
// @Tags Proxy
// @Accept json
// @Produce json
// @Param path path string true "Caminho da API da Binance (ex: /ticker/24hr)"
// @Success 200 {object} map[string]interface{}
// @Failure 502 {object} map[string]interface{}
// @Router /{path} [get]
// @Router /{path} [post]
func (p *ProxyServer) ProxyRequest(c *gin.Context) {
	// Tratar requisições OPTIONS (preflight CORS)
	if c.Request.Method == "OPTIONS" {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "3600")
		c.Status(http.StatusOK)
		return
	}

	// Obter o path da requisição (ex: /ticker/24hr, /klines, etc.)
	path := c.Request.URL.Path

	// Remover o prefixo /api se existir
	if strings.HasPrefix(path, "/api") {
		path = strings.TrimPrefix(path, "/api")
	}

	// Normalizar o path (garantir que comece com /)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Construir a URL completa da Binance
	targetURL := fmt.Sprintf("%s%s", p.binanceURL, path)

	// Processar query parameters e converter symbols se necessário
	queryParams := c.Request.URL.Query()
	if symbolsParam, exists := queryParams["symbols"]; exists && len(symbolsParam) > 0 {
		// A Binance espera symbols como array JSON: ["BTCUSDT","ETHUSDT"]
		// Mas pode vir como string separada por vírgulas: BTCUSDT,ETHUSDT
		symbolsValue := symbolsParam[0]
		
		// Verificar se já está no formato JSON array
		if !strings.HasPrefix(symbolsValue, "[") {
			// Converter string separada por vírgulas para array JSON
			symbolsList := strings.Split(symbolsValue, ",")
			// Limpar espaços em branco
			for i := range symbolsList {
				symbolsList[i] = strings.TrimSpace(symbolsList[i])
			}
			// Criar array JSON
			symbolsJSON, err := json.Marshal(symbolsList)
			if err == nil {
				// Substituir o valor do parâmetro
				queryParams.Set("symbols", string(symbolsJSON))
				// log.Printf("[DEBUG] Convertido symbols de '%s' para '%s'", symbolsValue, string(symbolsJSON))
			} else {
				// log.Printf("[WARN] Erro ao converter symbols para JSON: %v", err)
			}
		}
	}

	// Construir query string corrigida
	var queryString string
	if len(queryParams) > 0 {
		queryString = queryParams.Encode()
		targetURL += "?" + queryString
	}

	// log.Printf("[INFO] Proxying request: %s %s%s -> %s", c.Request.Method, c.Request.URL.Path, func() string {
	// 	if queryString != "" {
	// 		return "?" + queryString
	// 	} else if c.Request.URL.RawQuery != "" {
	// 		return "?" + c.Request.URL.RawQuery
	// 	}
	// 	return ""
	// }(), targetURL)

	// Criar a requisição para a Binance
	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1000,
			"msg":     fmt.Sprintf("Erro ao criar requisição: %v", err),
			"message": fmt.Sprintf("Erro ao criar requisição: %v", err),
		})
		return
	}

	// Copiar headers importantes (exceto Host e alguns headers problemáticos)
	for key, values := range c.Request.Header {
		keyLower := strings.ToLower(key)
		// Ignorar headers que não devem ser repassados
		if keyLower == "host" || keyLower == "connection" || keyLower == "keep-alive" {
			continue
		}
		// Modificar Accept-Encoding para evitar compressão desnecessária
		if keyLower == "accept-encoding" {
			req.Header.Set("Accept-Encoding", "gzip, deflate")
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Garantir que temos um User-Agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Binance-Proxy/1.0")
	}

	// Fazer a requisição para a Binance
	resp, err := p.client.Do(req)
	if err != nil {
		// log.Printf("Erro ao fazer requisição para Binance: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    -1000,
			"msg":     fmt.Sprintf("Erro ao conectar com Binance: %v", err),
			"message": fmt.Sprintf("Erro ao conectar com Binance: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	// Ler o corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// log.Printf("Erro ao ler resposta da Binance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1001,
			"msg":     "Erro ao ler resposta da Binance",
			"message": fmt.Sprintf("Erro ao ler resposta: %v", err),
		})
		return
	}

	// Log de debug do response
	// log.Printf("[DEBUG] Response Status: %d %s", resp.StatusCode, resp.Status)

	// Obter Content-Encoding para processamento
	contentEncoding := resp.Header.Get("Content-Encoding")

	// Se a resposta não for OK, logar o erro mas ainda processar o body
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// log.Printf("[WARN] Binance retornou status não-OK: %d %s", resp.StatusCode, resp.Status)
	}

	// Descomprimir se for gzip e preparar body para envio
	var bodyToSend []byte = body
	if contentEncoding == "gzip" {
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err == nil {
			decompressed, err := io.ReadAll(reader)
			reader.Close()
			if err == nil {
				bodyToSend = decompressed
				// log.Printf("[DEBUG] Descomprimido body gzip: %d bytes -> %d bytes", len(body), len(bodyToSend))
			} else {
				// log.Printf("[WARN] Erro ao descomprimir gzip: %v", err)
			}
		} else {
			// log.Printf("[WARN] Erro ao criar reader gzip: %v", err)
		}
	}

	// Tentar formatar o body como JSON para debug
	var jsonData interface{}
	if err := json.Unmarshal(bodyToSend, &jsonData); err == nil {
		// Se for JSON válido, formatar de forma legível (limitado a 2000 caracteres)
		prettyJSON, _ := json.MarshalIndent(jsonData, "", "  ")
		jsonStr := string(prettyJSON)
		if len(jsonStr) > 2000 {
			jsonStr = jsonStr[:2000] + "\n... (truncated)"
		}
		// log.Printf("[DEBUG] Response Body (JSON):\n%s", jsonStr)
	} else {
		// Se não for JSON, mostrar como string (limitado a 1000 caracteres)
		bodyStr := string(bodyToSend)
		if len(bodyStr) > 1000 {
			bodyStr = bodyStr[:1000] + "... (truncated)"
		}
		// log.Printf("[DEBUG] Response Body (raw): %s", bodyStr)
	}

	// Copiar headers importantes, mas remover Content-Encoding se descomprimimos
	for key, values := range resp.Header {
		keyLower := strings.ToLower(key)
		// Remover Content-Encoding se descomprimimos o body
		if keyLower == "content-encoding" && contentEncoding == "gzip" {
			continue
		}
		// Remover Content-Length pois pode mudar após descompressão
		if keyLower == "content-length" && contentEncoding == "gzip" {
			continue
		}
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Garantir que Content-Type esteja definido
	responseContentType := resp.Header.Get("Content-Type")
	if responseContentType == "" {
		// Tentar detectar o tipo de conteúdo baseado no body
		if len(bodyToSend) > 0 {
			var testJSON interface{}
			if json.Unmarshal(bodyToSend, &testJSON) == nil {
				responseContentType = "application/json"
			} else {
				responseContentType = "text/plain; charset=utf-8"
			}
		} else {
			responseContentType = "application/json"
		}
	}

	// Configurar CORS
	c.Header("Content-Type", responseContentType)
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Definir Content-Length correto
	if contentEncoding == "gzip" || c.Writer.Header().Get("Content-Length") == "" {
		c.Header("Content-Length", fmt.Sprintf("%d", len(bodyToSend)))
	}

	// Escrever status code e body
	c.Data(resp.StatusCode, responseContentType, bodyToSend)
}

// HealthCheck endpoint para verificar se o proxy está funcionando
// @Summary Health Check
// @Description Verifica se o proxy está funcionando corretamente
// @Tags Proxy
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (p *ProxyServer) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "ok",
		"service":    "binance-proxy",
		"time":       time.Now().Format(time.RFC3339),
		"binance_url": p.binanceURL,
	})
}

// TestConnection testa a conexão com a Binance
// @Summary Test Connection
// @Description Testa a conexão com a API da Binance
// @Tags Proxy
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /test [get]
func (p *ProxyServer) TestConnection(c *gin.Context) {
	testURL := fmt.Sprintf("%s/ping", p.binanceURL)

	resp, err := p.client.Get(testURL)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Erro ao conectar com Binance: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	c.JSON(http.StatusOK, gin.H{
		"status":      "ok",
		"binance_url": p.binanceURL,
		"http_status": resp.StatusCode,
		"message":     "Conexão com Binance estabelecida com sucesso",
	})
}

func setupRouter(proxy *ProxyServer) *gin.Engine {
	// Configurar Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Middleware CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "3600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// Rotas do proxy
	router.GET("/health", proxy.HealthCheck)
	router.GET("/test", proxy.TestConnection)

	// Handler customizado para Swagger que trata doc.json internamente
	swaggerHandler := func(c *gin.Context) {
		filepath := c.Param("filepath")
		
		// Se for doc.json, servir o JSON convertido do YAML
		if filepath == "/doc.json" || filepath == "doc.json" {
			// Ler o arquivo swagger.yaml
			yamlData, err := os.ReadFile("./swagger.yaml")
			if err != nil {
				// log.Printf("[ERROR] Erro ao ler swagger.yaml: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Não foi possível ler o arquivo swagger.yaml: %v", err),
				})
				return
			}

			// Converter YAML para JSON usando map[string]interface{} para melhor compatibilidade
			var swaggerData map[string]interface{}
			if err := yaml.Unmarshal(yamlData, &swaggerData); err != nil {
				// log.Printf("[ERROR] Erro ao converter YAML para JSON: %v", err)
				// log.Printf("[DEBUG] Primeiros 500 caracteres do YAML: %s", string(yamlData[:min(500, len(yamlData))]))
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Erro ao converter YAML para JSON: %v", err),
					"details": "Verifique os logs do servidor para mais informações",
				})
				return
			}

			// Validar se a conversão foi bem-sucedida
			if swaggerData == nil || len(swaggerData) == 0 {
				// log.Printf("[ERROR] YAML convertido está vazio")
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "YAML convertido está vazio",
				})
				return
			}

			c.JSON(http.StatusOK, swaggerData)
			return
		}

		// Para outros arquivos, usar o handler padrão do gin-swagger
		ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.URL("/swagger/doc.json"),
			ginSwagger.DefaultModelsExpandDepth(-1),
		)(c)
	}

	// Swagger UI - rota única com wildcard que trata tudo
	router.GET("/swagger/*filepath", swaggerHandler)

	// Proxy para todas as rotas da API da Binance (deve ser a última rota)
	router.NoRoute(proxy.ProxyRequest)

	return router
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
	router := setupRouter(proxy)

	// Configurar servidor HTTP
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// log.Printf("🚀 Proxy Binance iniciado na porta %s", port)
	// log.Printf("📡 URL da Binance: %s", binanceURL)
	// log.Printf("🌐 Endpoints disponíveis:")
	// log.Printf("   - GET  /health - Health check")
	// log.Printf("   - GET  /test - Testar conexão com Binance")
	// log.Printf("   - GET  /swagger/index.html - Documentação Swagger UI")
	// log.Printf("   - GET  /* - Proxy para API da Binance")
	// log.Printf("   - POST /* - Proxy para API da Binance")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
