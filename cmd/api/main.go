package main

import (
	"log"

	"github.com/NamespaceManager/internal/app"
)

// @title NamespaceManager API
// @version 1.0
// @description Resource Allocation and Management in Multi-Cluster project
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	db, err := app.InitDataBase()
	if err != nil {
		log.Fatal(err)
	}
	defer app.CloseDB()

	app := app.NewApp(db)
	if err := app.Run(); err != nil {
		log.Fatal("Server encountered an error:", err)
	}

}
