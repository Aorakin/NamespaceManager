package app

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/NamespaceManager/config"
	"github.com/NamespaceManager/docs"
	"github.com/NamespaceManager/pkg/httpclient"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

type App struct {
	gin        *gin.Engine
	postgresDB *gorm.DB
	// config      *config.Config
}

func NewApp(postgresDB *gorm.DB) *App {
	app := &App{
		gin:        gin.New(),
		postgresDB: postgresDB,
	}

	// Add request logging middleware
	app.gin.Use(gin.Logger())

	// Configure CORS
	app.gin.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Parse the origin URL to extract the host
			u, err := url.Parse(origin)
			if err != nil {
				return false
			}
			host := u.Host

			// Allow all subdomains of onepointfive.life
			return strings.HasSuffix(host, "onepointfive.life")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	return app
}

func (s *App) Run() error {
	err := s.gin.SetTrustedProxies([]string{"192.168.0.0/16", "10.0.0.0/8"})
	if err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	config.InitConfig()
	store := config.NewSessionStore("attc[pci7>klk-UQ!/h^b{!^rK{1mAe", 3600)
	s.gin.Use(sessions.Sessions("NSmanagerSession", store))

	docs.SwaggerInfo.Title = "ClearingHouse API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/"

	if err := s.MapHandlers(); err != nil {
		return err
	}

	// Initialize mTLS client for outbound requests to glidelet/resource controllers
	if err := httpclient.InitMTLSClient(); err != nil {
		log.Fatalf("Failed to initialize mTLS client: %v", err)
	}

	// Serve Swagger UI
	s.gin.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if os.Getenv("MTLS_ENABLED") == "true" {
		caCert, err := os.ReadFile(os.Getenv("MTLS_CA_CERT_PATH"))
		if err != nil {
			log.Fatalf("Failed to read CA certificate: %v", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			log.Fatal("Failed to parse CA certificate")
		}

		tlsConfig := &tls.Config{
			ClientCAs:  caCertPool,
			ClientAuth: tls.VerifyClientCertIfGiven,
			MinVersion: tls.VersionTLS12,
		}

		server := &http.Server{
			Addr:      ":8080",
			Handler:   s.gin,
			TLSConfig: tlsConfig,
		}

		log.Println("Starting server with mTLS support on :8080")
		return server.ListenAndServeTLS(
			os.Getenv("TLS_CERT_PATH"),
			os.Getenv("TLS_KEY_PATH"),
		)
	}

	serverURL := fmt.Sprintf(":%s", "8080")
	return s.gin.Run(serverURL)
}
