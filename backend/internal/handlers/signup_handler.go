package handlers

import (
	"errors"
	"fmt"
	"literary-lions/internal/models"
	"literary-lions/internal/utils"
	"literary-lions/internal/validator"
	"literary-lions/pkg/logger"
	"net/http"
)

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func UserSignup(w http.ResponseWriter, r *http.Request) error {
	logger.InfoLogger.Println("GET request for /signup")

	// Initialize the signup template
	tmpl, err := utils.InitTmpl("signup.html")
	if err != nil {
		return fmt.Errorf("failed to initialize signup.html: %w", err)
	}

	// Execute the signup template and send it to the client
	if err := tmpl.Execute(w, nil); err != nil {
		logger.ErrorLogger.Printf("Error executing template: %v", err)
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}

func renderSignupForm(w http.ResponseWriter, form userSignupForm) error {
	// Prepare data to be passed to the template
	data := map[string]interface{}{
		"FieldErrors": form.FieldErrors,
	}

	// Initialize the signup template
	tmpl, err := utils.InitTmpl("signup.html")
	if err != nil {
		return fmt.Errorf("failed to initialize signup.html: %w", err)
	}

	// Execute the signup template with the provided data
	if err := tmpl.Execute(w, data); err != nil {
		logger.ErrorLogger.Printf("Error executing template: %v", err)
		return fmt.Errorf("error executing template: %w", err)
	}
	return nil
}

func UserSignupPost(w http.ResponseWriter, r *http.Request) error {
	logger.InfoLogger.Println("POST request for /signup")

	// Parse the form data
	if err := r.ParseForm(); err != nil {
		logger.ErrorLogger.Println("Failed to parse form")
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return fmt.Errorf("failed to parse form: %w", err)
	}

	var form userSignupForm

	// Populate the form struct with the data from the request
	form.Name = r.FormValue("name")
	form.Email = r.FormValue("email")
	form.Password = r.FormValue("password")

	// Validate form fields
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "Enter a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "Password must be at least 8 characters long")

	if !form.Valid() {
		// Render the login form with errors
		return renderSignupForm(w, form)
	}

	// Create a new UserModel instance
	userModel, err := models.NewUserModel()
	if err != nil {
		logger.ErrorLogger.Printf("Failed to create user model: %v", err)
		return fmt.Errorf("failed to create user model: %w", err)
	}
	defer userModel.DB.Close()

	// Insert the new user into the database
	if err = userModel.Insert(form.Name, form.Email, form.Password); err != nil {
		logger.ErrorLogger.Printf("Error inserting user: %v", err)
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			return renderSignupForm(w, form)
		}
		if errors.Is(err, models.ErrDuplicateLogin) {
			form.AddFieldError("name", "Username is already in use")
			return renderSignupForm(w, form)
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	logger.InfoLogger.Printf("Successfully inserted user: %s", form.Name)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
	return nil
}
