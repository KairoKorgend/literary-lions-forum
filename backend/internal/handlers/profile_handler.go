package handlers

import (
	"fmt"
	"literary-lions/internal/models"
	"literary-lions/internal/utils"
	"literary-lions/pkg/logger"
	"net/http"
	"strconv"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request, userID string, sessionModel *models.SessionModel, userModel *models.UserModel, postModel *models.PostModel) error {
	logger.InfoLogger.Println("GET request for /profile")

	// Convert the requested user's id into an integer
	requestedUserID, err := strconv.Atoi(userID)
	if err != nil {
		logger.ErrorLogger.Printf("Invalid user ID format: %v", err)
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	// Get the post count for the requested user
	postCount, err := userModel.PostCount(requestedUserID)
	if err != nil {
		return fmt.Errorf("error getting post count: %w", err)
	}

	// Get the "tab" query parameter from the URL to determine which posts to fetch
	tab := r.URL.Query().Get("tab")
	if tab != "liked" {
		tab = "posts"
	}

	var posts []models.Post

	// Retrieve posts based on the tab value
	if tab == "posts" {
		// Get the requested user's posts
		posts, err = postModel.GetPostsByAuthorID(requestedUserID)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to get posts: %v", err)
			http.Error(w, "Failed to get posts", http.StatusInternalServerError)
			return fmt.Errorf("failed to get posts: %w", err)
		}
	} else {
		// Get the requested user's liked posts
		posts, err = postModel.GetLikedPostsByUserID(requestedUserID)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to get liked posts: %v", err)
			http.Error(w, "Failed to get liked posts", http.StatusInternalServerError)
			return fmt.Errorf("failed to get liked posts: %w", err)
		}
	}

	isAuthenticated, loggedInUserID, _ := sessionModel.IsAuthenticated(r)
	isOwnProfile := loggedInUserID == requestedUserID

	// Get the requested user's data
	user, err := userModel.Get(requestedUserID)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get user: %v", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return fmt.Errorf("failed to get user: %w", err)
	}

	data := map[string]interface{}{
		"User":            user,
		"Posts":           posts,
		"IsOwnProfile":    isOwnProfile,
		"IsAuthenticated": isAuthenticated,
		"loggedInUserID":  loggedInUserID,
		"ActiveTab":       tab,
		"PostCount":       postCount,
	}

	tmpl, err := utils.InitTmpl("profile.html")
	if err != nil {
		return fmt.Errorf("failed to initialize profile.html: %w", err)
	}

	if err := tmpl.Execute(w, data); err != nil {
		logger.ErrorLogger.Printf("Error executing template: %v", err)
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}
