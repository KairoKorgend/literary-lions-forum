package api

import (
	"database/sql"
	"literary-lions/internal/models"
	"literary-lions/pkg/logger"
	"net/http"
)

func StartAppServer(appPort string, staticDir string, db *sql.DB) error {
	// Create a new ServeMux for handling HTTP requests
	mux := http.NewServeMux()

	// Initialize models with the database connection
	sessionModel := &models.SessionModel{DB: db}
	userModel := &models.UserModel{DB: db}
	postModel := &models.PostModel{DB: db}
	categoryModel := &models.CategoryModel{DB: db}

	// Register routes with the ServeMux
	RegisterRoutes(mux, sessionModel, userModel, postModel, categoryModel, db)

	// Serve static files from the specified directory
	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	logger.InfoLogger.Println("Starting the application at port:", appPort)

	// Start the HTTP server
	if err := http.ListenAndServe(":"+appPort, mux); err != nil {
		logger.ErrorLogger.Println("failed to start application server:", err)
	}

	return nil
}
