package main

import (
	"os"

	"literary-lions/internal/api"
	"literary-lions/internal/config"
	dbserver "literary-lions/internal/db"
	"literary-lions/pkg/logger"
)

func main() {
	exists, err := dbserver.CheckDB()
	if err != nil {
		logger.ErrorLogger.Printf("Error checking database: %v", err)
		os.Exit(1)
	}

	if !exists {
		err = dbserver.CreateDB()
		if err != nil {
			logger.ErrorLogger.Printf("Error creating database: %v", err)
			os.Exit(1)
		}
		logger.InfoLogger.Println("Database Created")
	} else {
		logger.InfoLogger.Println("Database Exists")
	}

	db, err := dbserver.OpenDB()
	if err != nil {
		logger.ErrorLogger.Printf("Error opening database: %v", err)
		os.Exit(1)
	} else {
		logger.InfoLogger.Println("Successfully connected to database")
	}

	config, err := config.ReadConfig("../internal/config/app_server_config.json")
	if err != nil {
		logger.ErrorLogger.Printf("Failed to read configuration file: %v", err)
		os.Exit(1)
	}

	staticDir := "../../frontend/static"
	appPort := config.AppPort
	err = api.StartAppServer(appPort, staticDir, db)
	if err != nil {
		logger.ErrorLogger.Printf("Error starting application server: %v", err)
		os.Exit(1)
	}
	logger.InfoLogger.Println("Application Started")
}
