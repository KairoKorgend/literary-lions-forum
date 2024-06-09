package handlers

import (
	"fmt"
	"literary-lions/internal/models"
	"literary-lions/internal/utils"
	"literary-lions/pkg/logger"
	"net/http"
)

func PostHandler(w http.ResponseWriter, r *http.Request, sessionModel *models.SessionModel, userModel *models.UserModel, postModel *models.PostModel) error {
	// Get required data from the URL
	postID, showModal, action, commentID, err := utils.ParseURL(r)
	if err != nil {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return fmt.Errorf("failed to parse URL path: %w", err)
	}

	// Get the requested post by the postID
	post, err := postModel.GetPostByID(postID)
	if err != nil {
		return fmt.Errorf("failed to retrieve post: %w", err)
	}

	if post == nil {
		http.NotFound(w, r)
		return nil
	}

	var replyTo interface{} = nil

	if action == "reply" {
		replyTo = commentID
	}

	tmpl, err := utils.InitTmpl("postPage.html")
	if err != nil {
		return fmt.Errorf("failed to initialize postPage.html: %w", err)
	}

	// Get the post count for the author
	postCount, err := userModel.PostCount(post.AuthorID)
	if err != nil {
		return fmt.Errorf("error getting post count: %w", err)
	}

	// Get the creation date of the author's account
	userSince, err := userModel.GetUserCreationDateByAuthorID(post.AuthorID)
	if err != nil {
		return fmt.Errorf("error getting author's user creation date: %w", err)
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

		// Get author's data
		authorData, err := userModel.Get(loggedInUserID)
		if err != nil {
			return fmt.Errorf("error retrieving user information: %w", err)
		}

		// Get the user's reactions to the post
		err = postModel.GetUserPostReactions(post, loggedInUserID)
		if err != nil {
			return fmt.Errorf("error retrieving user reactions: %w", err)
		}

		// Get the user's reactions to the comments
		for i := range post.Comments {
			err = postModel.GetUserCommentReactions(&post.Comments[i], loggedInUserID)
			if err != nil {
				return fmt.Errorf("error retrieving user reactions: %w", err)
			}
		}

		// Determine if the post belongs to the user
		isOwnPost := post.AuthorID == loggedInUserID

		if isOwnPost && action == "delete" {
			err := postModel.DeletePost(postID)
			if err != nil {
				return fmt.Errorf("error deleting post: %w", err)
			}
			logger.InfoLogger.Printf("Post with ID %d deleted by user %d", postID, loggedInUserID)
			http.Redirect(w, r, "/", http.StatusFound)
			return nil
		}

		if r.Method == http.MethodPost {
			switch action {
			case "comment":
				// Comment under the post
				err := postModel.SubmitComment(r, postModel, postID, loggedInUserID, authorData.Login)
				if err != nil {
					return fmt.Errorf("error submitting comment: %w", err)
				}
				logger.InfoLogger.Printf("User %d submitted a comment on post %d", loggedInUserID, postID)

			case "reply":
				// Reply to a comment under the post
				err := postModel.SubmitReply(r, postModel, commentID, loggedInUserID, authorData.Login)
				if err != nil {
					return fmt.Errorf("error submitting reply: %w", err)
				}
				logger.InfoLogger.Printf("User %d submitted a reply to comment %d on post %d", loggedInUserID, commentID, postID)
			}
			http.Redirect(w, r, fmt.Sprintf("/postPage/%d", postID), http.StatusFound)
		}

		data = map[string]interface{}{
			"IsAuthenticated": isAuthenticated,
			"UserID":          loggedInUserID,
			"IsOwnPost":       isOwnPost,
			"PostCount":       postCount,
			"User":            authorData,
			"UserSince":       userSince,
			"Post":            post,
			"ShowModal":       showModal,
			"ReplyTo":         replyTo,
		}
	} else {

		data = map[string]interface{}{
			"IsAuthenticated": isAuthenticated,
			"Post":            post,
			"PostCount":       postCount,
			"UserSince":       userSince,
		}
	}

	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	logger.InfoLogger.Printf("GET request for /postpage/%d", postID)

	return nil
}
