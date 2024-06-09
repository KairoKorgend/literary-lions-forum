package handlers

import (
	"fmt"
	"literary-lions/internal/models"
	"literary-lions/pkg/logger"
	"net/http"
	"strconv"
)

func ReactHandler(w http.ResponseWriter, r *http.Request, postModel *models.PostModel) error {

	var table string
	var column string
	var id int

	// Parse form data from the request
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return fmt.Errorf("failed to parse form data: %w", err)
	}

	// Determine the action (like or dislike) and set the appropriate table
	action := r.Form.Get("action")
	if action == "like" {
		table = "likes"
	} else if action == "dislike" {
		table = "dislikes"
	} else {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return fmt.Errorf("invalid form action: %s", action)
	}

	// Get and validate the reacting user's ID
	userIDStr := r.Form.Get("reacting_user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return fmt.Errorf("invalid user id: %w", err)
	}

	postIDstr := r.Form.Get("post_id")
	postID, err := strconv.Atoi(postIDstr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return fmt.Errorf("invalid post id: %w", err)
	}

	// Determine if the reaction is for a post or a comment and set the appropriate column and ID
	if r.Form.Get("type") == "post" {
		column = "post_id"
		id, _ = strconv.Atoi(r.Form.Get("post_id"))
	} else if r.Form.Get("type") == "comment" {
		column = "comment_id"
		id, _ = strconv.Atoi(r.Form.Get("comment_id"))
	} else {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return fmt.Errorf("invalid form data: %w", err)
	}

	// Insert the reaction into the database
	err = postModel.InsertReaction(table, column, id, userID, true)
	if err != nil {
		return fmt.Errorf("failed to insert reaction: %w", err)
	}

	logger.InfoLogger.Printf("User %d %sd a %s with ID %d", userID, action, r.Form.Get("type"), id)

	http.Redirect(w, r, fmt.Sprintf("/postPage/%d", postID), http.StatusFound)

	return nil
}
