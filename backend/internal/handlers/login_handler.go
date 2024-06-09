package handlers

import (
	"errors"
	"fmt"
	"literary-lions/internal/models"
	"literary-lions/internal/utils"
	"literary-lions/internal/validator"
	"literary-lions/pkg/logger"
	"net/http"

	"github.com/google/uuid"
)

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func UserLoginGet(w http.ResponseWriter, r *http.Request) error {
	logger.InfoLogger.Println("GET request for /login")

	// Initialize the login template
	tmpl, err := utils.InitTmpl("login.html")
	if err != nil {
		return fmt.Errorf("failed to initialize login.html: %w", err)
	}

	form := userLoginForm{}
	data := map[string]interface{}{
		"Form": form,
	}

	// Execute the login template and send it to the client
	if err := tmpl.Execute(w, data); err != nil {
		logger.ErrorLogger.Printf("Error executing template: %v", err)
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}

func renderLoginForm(w http.ResponseWriter, form userLoginForm) error {
	// Prepare data to be passed to the template
	data := map[string]interface{}{
		"FieldErrors":    form.FieldErrors,
		"Form":           form,
		"NonFieldErrors": form.NonFieldErrors,
	}
	// Initialize the login template
	tmpl, err := utils.InitTmpl("login.html")
	if err != nil {
		return fmt.Errorf("failed to initialize login.html: %w", err)
	}

	// Execute the login template with the provided data
	if err := tmpl.Execute(w, data); err != nil {
		logger.ErrorLogger.Printf("Error executing template: %v", err)
		return fmt.Errorf("error executing template: %w", err)
	}
	return nil
}

func UserLoginPost(w http.ResponseWriter, r *http.Request) error {
	logger.InfoLogger.Println("POST request for /login")

	// Parse the form data
	if err := r.ParseForm(); err != nil {
		logger.ErrorLogger.Printf("Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return fmt.Errorf("failed to parse form: %w", err)
	}

	var form userLoginForm

	// Populate the form struct with the data from the request
	form.Email = r.FormValue("email")
	form.Password = r.FormValue("password")

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		// Render the login form with errors
		return renderLoginForm(w, form)
	}

	// Create a new UserModel instance
	userModel, err := models.NewUserModel()
	if err != nil {
		return fmt.Errorf("failed to create user model: %w", err)
	}
	defer userModel.DB.Close()

	// Authenticate the user
	user, err := userModel.Authenticate(form.Email, form.Password)
	if err != nil {
		logger.ErrorLogger.Printf("Error authenticating user: %v", err)
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			// Render the signup form with errors
			return renderLoginForm(w, form)
		}
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Generate a new UUID for the session
	sessionID := uuid.New()

	// Create a new SessionModel instance
	sessionModel, err := models.NewSessionModel()
	if err != nil {
		return fmt.Errorf("failed to create session model: %w", err)
	}
	defer sessionModel.DB.Close()

	// Create a new session for the user
	if err := sessionModel.Create(w, user.ID, sessionID); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Log that the session has started
	logger.InfoLogger.Printf("Session started for user with ID: %d | session ID: %s", user.ID, sessionID)

	// Redirect to the main page
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}
