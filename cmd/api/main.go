package main

import (
	"log"

	"github.com/NamespaceManager/internal/app"
)

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
