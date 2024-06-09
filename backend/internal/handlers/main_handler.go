package handlers

import (
	"fmt"
	"literary-lions/internal/models"
	"literary-lions/internal/utils"
	"literary-lions/internal/validator"
	"literary-lions/pkg/logger"
	"net/http"
)

func MainHandler(w http.ResponseWriter, r *http.Request, sessionModel *models.SessionModel, userModel *models.UserModel, subcategoryId int, searchString string, filter string) error {
	if r.URL.Path == "/" {
        logger.InfoLogger.Println("GET request for /")
    }
	
	tmpl, err := utils.InitTmpl("index.html")
	if err != nil {
		return fmt.Errorf("failed to initialize index.html: %w", err)
	}

	// Create a new CategoryModel instance
	categoryModel, err := models.NewCategoryModel()
	if err != nil {
		return fmt.Errorf("failed to create new category model: %w", err)
	}

	// Get category data
	categories, err := categoryModel.Get()
	if err != nil {
		return fmt.Errorf("failed to retrieve categories: %w", err)
	}

	// Create a new PostModel instance
	postsModel, err := models.NewPostModel()
	if err != nil {
		return fmt.Errorf("failed to create new posts model: %w", err)
	}

	err = validator.ValidateSearchString(searchString)
	if err != nil {
		logger.ErrorLogger.Printf("Invalid search parameter: %v", err)
		searchString = ""
	}

	// Get posts data based on the provided queries
	queryString, queryParams := utils.IndexGetPostsQueryBuilder(subcategoryId, searchString, filter)
	posts, err := postsModel.Get(queryString, queryParams...)
	if err != nil {
		http.Error(w, "Failed to retrieve posts: %v", http.StatusNotFound)
		return fmt.Errorf("failed to retrieve posts: %w", err)
	}

	// Log the selected filter if one is selected
	if filter != "" {
    logger.InfoLogger.Printf("Posts filtered using: %s", filter)
	}

	// Check if the user is authenticated
	isAuthenticated, loggedInUserID, err := sessionModel.IsAuthenticated(r)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to check if user is authenticated: %v", err)
		http.Error(w, "Failed to check if the user is authenticated: %w", http.StatusBadRequest)
	}

	var data map[string]interface{}

	// Pass in specific data depending if the user is authenticated or not
	if isAuthenticated {

		// Get user data
		user, err := userModel.Get(loggedInUserID)
		if err != nil {
			return fmt.Errorf("error retrieving user information: %w", err)
		}

		// Get the user's post reactions data
		for i := range posts {
			err := postsModel.GetUserPostReactions(&posts[i], loggedInUserID)
			if err != nil {
				return fmt.Errorf("error retrieving user reactions")
			}
		}

		data = map[string]interface{}{
			"IsAuthenticated": isAuthenticated,
			"UserID":          loggedInUserID,
			"User":            user,
			"Categories":      categories,
			"Posts":           posts,
		}
	} else {
		data = map[string]interface{}{
			"IsAuthenticated": isAuthenticated,
			"Categories":      categories,
			"Posts":           posts,
		}
	}
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}
