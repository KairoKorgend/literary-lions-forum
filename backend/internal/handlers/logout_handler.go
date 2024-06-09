package handlers

import (
	"fmt"
	"literary-lions/internal/models"
	"literary-lions/pkg/logger"
	"net/http"

	"github.com/google/uuid"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request, sessionModel *models.SessionModel) error {
	logger.InfoLogger.Println("POST request to logout from /profile")

	// Get the session ID from the cookie
	cookie, err := r.Cookie("session")
	if err != nil {
		return fmt.Errorf("failed to get session ID: %w", err)
	}

	// Parse the session ID from the cookie
	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		return fmt.Errorf("failed to parse session ID: %w", err)
	}

	// Check if the user is authenticated
	isAuthenticated, _, err := sessionModel.IsAuthenticated(r)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to check if user is authenticated: %v", err)
		http.Error(w, "Failed to check if the user is authenticated: %w", http.StatusBadRequest)
	}

	// Delete the session, logging the user out
	if isAuthenticated {
		err := sessionModel.Delete(sessionID)
		if err != nil {
			return fmt.Errorf("failed to log out: %w", err)
		}

		logger.InfoLogger.Printf("User with session ID %s has been logged out.", sessionID)
	}

	// Redirect the user to the home page after logging out
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}
