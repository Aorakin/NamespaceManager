package app

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/NamespaceManager/config"
	"github.com/NamespaceManager/docs"
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

	// Do not expose Swagger docs in production.
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if env != "prod" && env != "production" {
		s.gin.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	serverURL := fmt.Sprintf(":%s", "8080")
	return s.gin.Run(serverURL)
}
