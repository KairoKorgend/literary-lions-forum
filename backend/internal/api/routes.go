package api

import (
	"database/sql"
	"literary-lions/internal/handlers"
	"literary-lions/internal/models"
	"literary-lions/pkg/logger"
	"net/http"
	"strconv"
	"strings"
)

func RegisterRoutes(mux *http.ServeMux, sessionModel *models.SessionModel, userModel *models.UserModel, postModel *models.PostModel, categoryModel *models.CategoryModel, db *sql.DB) error {
	logger.InfoLogger.Println("Registering routes")

	// Route for the main page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		if r.Method == "GET" {
			filter := r.URL.Query().Get("filter")
			searchString := r.URL.Query().Get("search")

			// Handle the main page request
			err := handlers.MainHandler(w, r, sessionModel, userModel, 0, searchString, filter)
			if err != nil {
				http.Error(w, "Failed to handle main page", http.StatusInternalServerError)
				logger.ErrorLogger.Printf("Failed to handle main page: %v", err)
				return
			}
		}
	})

	// Route for subcategory pages
	mux.HandleFunc("/subcategory/", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) >= 2 {
			subcategoryID, err := strconv.Atoi(pathParts[1])
			if err != nil {
				logger.ErrorLogger.Printf("Invalid subcategory ID format: %v", err)
				http.Error(w, "Invalid subcategory ID format", http.StatusBadRequest)
			}

			// Hande the subcategory page request
			err = handlers.MainHandler(w, r, sessionModel, userModel, subcategoryID, "", filter)
			if err != nil {
				http.Error(w, "Failed to handle main page", http.StatusInternalServerError)
				logger.ErrorLogger.Printf("Failed to handle main page: %v", err)
				return
			}
		}
	})

	// Route for the profile pages
	mux.HandleFunc("/profile/", func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 2 {
			http.Error(w, "Invalid profile URL", http.StatusBadRequest)
			return
		}

		userID := pathParts[1]

		if _, err := strconv.Atoi(userID); err != nil {
			logger.ErrorLogger.Printf("Invalid user ID format: %v", err)
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}

		// Check if the URL is for changing the password
		if len(pathParts) == 3 && pathParts[2] == "password" {
			// Handle the password change request
			err := handlers.ChangePasswordHandler(w, r, userID, userModel)
			if err != nil {
				http.Error(w, "Failed to handle password change", http.StatusInternalServerError)
				logger.ErrorLogger.Printf("Failed to handle password change: %v", err)
				return
			}
		} else {
			// Handle the profile page request
			err := handlers.ProfileHandler(w, r, userID, sessionModel, userModel, postModel)
			if err != nil {
				http.Error(w, "Failed to handle profile page", http.StatusInternalServerError)
				logger.ErrorLogger.Printf("Failed to handle profile page: %v", err)
				return
			}
		}
	})

	// Route for image uploads
	mux.HandleFunc("/upload-image", func(w http.ResponseWriter, r *http.Request) {
		// Handle the image upload request
		handlers.UploadImageHandler(w, r, sessionModel, db)
	})

	// Route for creating posts
	mux.HandleFunc("/createPost", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Handle the GET request for creating a post
			err := handlers.CreatePostHandlerGet(w, r, sessionModel, userModel, categoryModel)
			if err != nil {
				http.Error(w, "Failed to handle post creating page", http.StatusInternalServerError)
				logger.ErrorLogger.Printf("Failed to handle post creating page: %v", err)
				return
			}
		case http.MethodPost:
			// Handle the POST request for creating a post
			err := handlers.CreatePostHandlerPost(w, r, sessionModel, userModel, postModel)
			if err != nil {
				http.Error(w, "Failed to handle post creating page", http.StatusInternalServerError)
				logger.ErrorLogger.Printf("Failed to handle post creating page: %v", err)
				return
			}
		}
	})

	// Route for viewing a single post
	mux.HandleFunc("/postPage/", func(w http.ResponseWriter, r *http.Request) {
		// Handle the post page request
		err := handlers.PostHandler(w, r, sessionModel, userModel, postModel)
		if err != nil {
			http.Error(w, "Failed to handle post page", http.StatusInternalServerError)
			logger.ErrorLogger.Printf("Failed to handle post page: %v", err)
			return
		}
	})

	// Route for logging out
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		// Handle the logout request
		handlers.LogoutHandler(w, r, sessionModel)
	})

	// Route for logging in
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Handle the GET request for login
			err := handlers.UserLoginGet(w, r)
			if err != nil {
				http.Error(w, "Failed to handle GET /login", http.StatusInternalServerError)
				logger.ErrorLogger.Printf("Failed to handle GET /login: %v", err)
				return
			}
		case http.MethodPost:
			// Handle the POST request for login
			err := handlers.UserLoginPost(w, r)
			if err != nil {
				http.Error(w, "Failed to handle POST /login", http.StatusInternalServerError)
				logger.ErrorLogger.Printf("Failed to handle POST /login: %v", err)
				return
			}
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	// Route for signing up
	mux.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Handle the GET request for signup
			err := handlers.UserSignup(w, r)
			if err != nil {
				http.Error(w, "Failed to handle GET /signup", http.StatusInternalServerError)
				logger.ErrorLogger.Printf("Failed to handle GET /signup: %v", err)
				return
			}
		case http.MethodPost:
			// Handle the POST request for signup
			err := handlers.UserSignupPost(w, r)
			if err != nil {
				http.Error(w, "Failed to handle POST /signup", http.StatusInternalServerError)
				logger.ErrorLogger.Printf("Failed to handle POST /signup: %v", err)
				return
			}
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	// Route for reacting to posts/comments
	mux.HandleFunc("/react", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// Handle the POST request for reactions
			err := handlers.ReactHandler(w, r, postModel)
			if err != nil {
				http.Error(w, "Failed to handle POST /react", http.StatusInternalServerError)
				logger.ErrorLogger.Printf("Failed to handle POST /react: %v", err)
				return
			}
		}

	})

	return nil
}
