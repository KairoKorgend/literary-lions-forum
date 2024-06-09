package handlers

import (
	"errors"
	"fmt"
	"literary-lions/internal/models"
	"literary-lions/internal/utils"
	"literary-lions/pkg/logger"
	"net/http"
	"strconv"
)

type TemplateData struct {
	UserID int
	Error  string
}

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request, userID string, userModel *models.UserModel) error {
	logger.InfoLogger.Printf("Password change request from /profile for user ID %s", userID)

	// Convert userID to int
	id, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return fmt.Errorf("invalid user ID format")
	}
	tmpl, err := utils.InitTmpl("password.html")
	if err != nil {
		return fmt.Errorf("failed to initialize password.html: %w", err)
	}

	switch r.Method {
	case http.MethodGet:
		// Create a data object with the user ID
		data := TemplateData{
			UserID: id,
		}

		// Pass the data object to the template
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to execute template", http.StatusInternalServerError)
			return fmt.Errorf("failed to execute template: %w", err)
		}

	case http.MethodPost:
		// Parse the form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return fmt.Errorf("failed to parse form: %w", err)
		}

		// Get the form values
		currentPassword := r.FormValue("currentPassword")
		newPassword := r.FormValue("newPassword")
		newPasswordConfirmation := r.FormValue("newPasswordConfirmation")

		// Validate the form values
		if newPassword != newPasswordConfirmation {
			// Create a data object with the error message
			data := TemplateData{
				UserID: id,
				Error:  "New passwords do not match",
			}

			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Failed to execute template", http.StatusInternalServerError)
				return fmt.Errorf("failed to execute template: %w", err)
			}

			return errors.New("new passwords do not match")
		}

		// Update the user's password
		if err := userModel.PasswordUpdate(id, currentPassword, newPassword); err != nil {
			// Create a data object with the error message
			data := TemplateData{
				UserID: id,
			}

			if errors.Is(err, models.ErrInvalidCredentials) {
				logger.ErrorLogger.Println("Current password is incorrect")
				data.Error = "Current password is incorrect"
			} else {
				logger.ErrorLogger.Println("Failed to update password")
				data.Error = "Failed to update password"
			}

			// Render the password change form with the error message
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Failed to execute template", http.StatusInternalServerError)
				return fmt.Errorf("failed to execute template: %w", err)
			}

			return err
		}

		// Log a message indicating that the password change was successful
		logger.InfoLogger.Printf("Password change successful for user ID %s", userID)

		// Redirect the user to their profile page
		http.Redirect(w, r, "/profile/"+userID, http.StatusSeeOther)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return errors.New("method not allowed")
	}

	return nil
}
