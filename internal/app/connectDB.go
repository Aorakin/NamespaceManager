package app

import (
	"fmt"
	"log"
	"os"

	"github.com/NamespaceManager/internal/models"
	// "github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDataBase() (*gorm.DB, error) {
	// err := godotenv.Load()
	// if err != nil {
	// 	return nil, fmt.Errorf("Error loading .env file :%v", err)
	// }
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	DB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Enable logging
	})
	if err != nil {
		return nil, fmt.Errorf("could not connect to the database: %v", err)
	}

	// Drop old tables that might conflict
	// DB.Migrator().DropTable("glider_specs", "glider_tickets", "tasks")

	// DB.Migrator().DropTable("tickets", "tasks")

	// Migrate with error handling
	err = DB.AutoMigrate(&models.User{}, &models.UserProvider{}, &models.Task{}, &models.Ticket{}, &models.QueueTicket{})
	if err != nil {
		return nil, fmt.Errorf("could not migrate database: %v", err)
	}

	log.Println("Database migration completed successfully")
	return DB, nil
}

func CloseDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("error getting database connection: %v", err)
	}
	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("error closing the database connection: %v", err)
	}

	log.Println("Database connection closed.")
	return nil
}
