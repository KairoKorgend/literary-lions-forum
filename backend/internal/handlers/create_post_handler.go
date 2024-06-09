package handlers

import (
	"fmt"
	"literary-lions/internal/models"
	"literary-lions/internal/utils"
	"literary-lions/pkg/logger"
	"net/http"
	"strconv"
)

type createPostForm struct {
	Title         string `form:"title"`
	Content       string `form:"content"`
	SubcategoryID int    `form:"subcategory"`
}

func CreatePostHandlerGet(w http.ResponseWriter, r *http.Request, sessionModel *models.SessionModel, userModel *models.UserModel, categoryModel *models.CategoryModel) error {
	logger.InfoLogger.Println("GET request for /createPost")

	// Check if the user is authenticated
	isAuthenticated, loggedInUserID, err := sessionModel.IsAuthenticated(r)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to check if user is authenticated: %v", err)
		http.Error(w, "Failed to check if the user is authenticated: %w", http.StatusBadRequest)
	}

	if !isAuthenticated {
		http.Error(w, "You need to be logged in to view this page", http.StatusUnauthorized)
	}

	// Fetch the user's details from the database
	user, err := userModel.Get(loggedInUserID)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get user: %v", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return fmt.Errorf("failed to get user: %w", err)
	}
	categories, err := categoryModel.Get()
	if err != nil {
		return fmt.Errorf("failed to retrieve categories: %w", err)
	}
	var data map[string]interface{}

	// Format the user's creation date
	user.CreatedAtFormatted = user.Created.Format("02.01.2006")

	tmpl, err := utils.InitTmpl("createPost.html")
	if err != nil {
		return fmt.Errorf("failed to initialize createPost.html: %w", err)
	}

	data = map[string]interface{}{
		"User":       user,
		"Categories": categories,
	}

	// Execute the template and pass the user to it
	if err := tmpl.Execute(w, data); err != nil {
		logger.ErrorLogger.Printf("Error executing template: %v", err)
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}

func CreatePostHandlerPost(w http.ResponseWriter, r *http.Request, sessionModel *models.SessionModel, userModel *models.UserModel, postModel *models.PostModel) error {
	logger.InfoLogger.Println("POST request from /createPost")

	// Parse and validate the form data
	if err := r.ParseForm(); err != nil {
		logger.ErrorLogger.Printf("Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return fmt.Errorf("failed to parse form: %w", err)
	}

	var form createPostForm

	// Populate the form struct with the data from the request
	form.Title = r.FormValue("title")
	form.Content = r.FormValue("content")
	var err error
	form.SubcategoryID, err = strconv.Atoi(r.FormValue("subcategory"))
	if err != nil {
		logger.ErrorLogger.Printf("Error converting subcategoryID: %v", err)
		http.Redirect(w, r, fmt.Sprintln("/createPost"), http.StatusSeeOther)
		return fmt.Errorf("error converting subcategoryID: %w", err)
	}

	// Check if form inputs are empty
	if form.Title == "" || form.Content == "" || form.SubcategoryID == 0 {
		logger.ErrorLogger.Print("Title and content cannot be empty", http.StatusBadRequest)
		http.Redirect(w, r, fmt.Sprintln("/createPost"), http.StatusSeeOther)
		return fmt.Errorf("title and content cannot be empty")
	}

	// Get the user's ID from the session
	userID, err := sessionModel.GetUserID(r)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get user id: %v", err)
		http.Error(w, "Failed to get user id", http.StatusInternalServerError)
		return fmt.Errorf("failed to get user id: %w", err)
	}

	// Get the user's login from the UserModel
	user, err := userModel.Get(userID)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get user: %v", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Insert the new post into the database
	postID, err := postModel.CreatePost(form.Title, form.Content, userID, form.SubcategoryID, user.Login)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	logger.InfoLogger.Printf("Successfully created new post with ID:%d", postID)
	http.Redirect(w, r, fmt.Sprintf("/postPage/%d", postID), http.StatusSeeOther)
	return nil
}
