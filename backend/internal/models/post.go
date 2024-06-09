package models

import (
	"database/sql"
	"fmt"
	"html/template"
	dbserver "literary-lions/internal/db"
	"literary-lions/internal/utils"
	"literary-lions/pkg/logger"
	"net/http"
	"strings"
	"time"
)

type Post struct {
	ID            int
	Title         string
	Content       template.HTML
	AuthorName    string
	AuthorID      int
	AuthorImage   string
	SubcategoryID int
	EventTime     time.Time
	EventTimeAgo  string
	Likes         int
	Dislikes      int
	HasLiked      bool
	HasDisliked   bool
	Comments      []Comment
	Subcategory   Subcategory
}

type Comment struct {
	ID          int
	PostID      int
	AuthorID    int
	AuthorName  string
	Content     template.HTML
	AuthorImage string
	CreatedAt   time.Time
	Likes       int
	Dislikes    int
	HasLiked    bool
	HasDisliked bool
	Replies     []Reply
}

type Reply struct {
	ID          int
	CommentID   int
	AuthorID    int
	AuthorName  string
	AuthorImage string
	Content     template.HTML
	Image       string
	CreatedAt   time.Time
}

type PostModel struct {
	DB *sql.DB
}

func NewPostModel() (*PostModel, error) {
	db, err := sql.Open("sqlite3", DatabaseLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &PostModel{DB: db}, nil
}

func (m *PostModel) CreatePost(title string, content string, authorID int, subcategoryID int, userName string) (int, error) {
	sanitizedContent := utils.ParseAndReplaceImageTags(content)

	// SQL statement to insert a new post
	stmt := `INSERT INTO posts (title, content, author_id, author, subcategory_id, event_time) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`
	result, err := m.DB.Exec(stmt, title, sanitizedContent, authorID, userName, subcategoryID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert post: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (m *PostModel) DeletePost(postID int) error {
    tx, err := m.DB.Begin()
    if err != nil {
        logger.ErrorLogger.Printf("Failed to begin transaction: %v", err)
        return err
    }

    // Delete replies
    _, err = tx.Exec("DELETE FROM replies WHERE comment_id IN (SELECT id FROM comments WHERE post_id = ?)", postID)
    if err != nil {
        logger.ErrorLogger.Printf("Failed to delete replies: %v", err)
        tx.Rollback()
        return fmt.Errorf("failed to delete replies: %w", err)
    }

    // Delete comments
    _, err = tx.Exec("DELETE FROM comments WHERE post_id = ?", postID)
    if err != nil {
        logger.ErrorLogger.Printf("Failed to delete comments: %v", err)
        tx.Rollback()
        return fmt.Errorf("failed to delete comments: %w", err)
    }

    // Delete post
    _, err = tx.Exec("DELETE FROM posts WHERE id = ?", postID)
    if err != nil {
        logger.ErrorLogger.Printf("Failed to delete post: %v", err)
        tx.Rollback()
        return fmt.Errorf("failed to delete post: %w", err)
    }

    // Delete likes associated with the post
    _, err = tx.Exec("DELETE FROM likes WHERE post_id = ?", postID)
    if err != nil {
        logger.ErrorLogger.Printf("Failed to delete likes: %v", err)
        tx.Rollback()
        return fmt.Errorf("failed to delete likes: %w", err)
    }

    // Delete dislikes associated with the post
    _, err = tx.Exec("DELETE FROM dislikes WHERE post_id = ?", postID)
    if err != nil {
        logger.ErrorLogger.Printf("Failed to delete dislikes: %v", err)
        tx.Rollback()
        return fmt.Errorf("failed to delete dislikes: %w", err)
    }

    logger.InfoLogger.Printf("Deleted post with post ID %d", postID)

    err = tx.Commit()
    if err != nil {
        logger.ErrorLogger.Printf("Failed to commit transaction: %v", err)
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

func (pm *PostModel) Get(queryString string, args ...interface{}) ([]Post, error) {
	// Execute the query to retrieve posts
	rows, err := pm.DB.Query(queryString, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	posts := []Post{}

	// Iterate over the rows returned by the query
	for rows.Next() {
		var post Post
		if err := rows.Scan(
			&post.ID,
			&post.SubcategoryID,
			&post.AuthorName,
			&post.AuthorID,
			&post.AuthorImage,
			&post.EventTime,
			&post.Title,
			&post.Content,
			&post.Likes,
			&post.Dislikes); err != nil {
			return nil, fmt.Errorf("failed row scan: %w", err)
		}

		// Convert event time to a human-readable format
		post.EventTimeAgo = utils.TimeAgo(post.EventTime)

		// Get the title of the subcategory
		title, err := pm.getSubcategoryTitle(post.SubcategoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get subcategory title: %w", err)
		}
		post.Subcategory = Subcategory{ID: post.SubcategoryID, Title: title}

		// Get comments associated with the post
		comments, err := pm.GetComments(post.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve comments: %w", err)
		}
		post.Comments = comments
		posts = append(posts, post)
	}

	return posts, nil
}

func (pm *PostModel) GetPostByID(postID int) (*Post, error) {
	// Execute the query to retrieve the post
	row := pm.DB.QueryRow(dbserver.SinglePostQuery, postID)
	var post Post
	if err := row.Scan(&post.ID, &post.SubcategoryID, &post.AuthorName, &post.AuthorID, &post.AuthorImage,
		&post.EventTime, &post.Title, &post.Content, &post.Likes, &post.Dislikes); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No post found
		}
		return nil, fmt.Errorf("failed to retrieve post: %w", err)
	}
	// Get the title of the subcategory
	title, err := pm.getSubcategoryTitle(post.SubcategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subcategory title: %w", err)
	}
	post.Subcategory = Subcategory{ID: post.SubcategoryID, Title: title}

	// Get comments associated with the post
	comments, err := pm.GetComments(post.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve comments: %w", err)
	}
	post.Comments = comments

	return &post, nil
}

func (pm *PostModel) GetPostsByAuthorID(authorID int) ([]Post, error) {
	// Execute the query to retrieve posts
	stmt := strings.Split(dbserver.IndexPostQuery, "ORDER BY")[0] + " WHERE p.author_id = ?"
	rows, err := pm.DB.Query(stmt, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve posts: %w", err)
	}
	defer rows.Close()

	var posts []Post

	// Iterate over the rows returned by the query
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.SubcategoryID, &post.AuthorName, &post.AuthorID, &post.AuthorImage, &post.EventTime, &post.Title, &post.Content, &post.Likes, &post.Dislikes); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		// Convert event time to a human-readable format
		post.EventTimeAgo = utils.TimeAgo(post.EventTime)

		// Get the title of the subcategory
		title, err := pm.getSubcategoryTitle(post.SubcategoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get subcategory title: %w", err)
		}
		post.Subcategory = Subcategory{ID: post.SubcategoryID, Title: title}

		// Get comments associated with the post
		comments, err := pm.GetComments(post.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve comments: %w", err)
		}
		post.Comments = comments

		posts = append(posts, post)
	}

	return posts, nil
}
func (pm *PostModel) GetLikedPostsByUserID(userID int) ([]Post, error) {
	// Execute the query to retrieve liked posts
	rows, err := pm.DB.Query(dbserver.LikedPostsQuery, userID) // Pass userID to filter the posts
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve liked posts: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		// Ensure your SELECT list in the IndexPostQuery aligns with the fields being scanned here
		if err := rows.Scan(&post.ID, &post.SubcategoryID, &post.AuthorName, &post.AuthorID, &post.AuthorImage, &post.EventTime, &post.Title, &post.Content, &post.Likes, &post.Dislikes); err != nil {
			return nil, fmt.Errorf("failed to scan liked post: %w", err)
		}

		// Convert event time to a human-readable format
		post.EventTimeAgo = utils.TimeAgo(post.EventTime)

		// Fetch the Subcategory title
		title, err := pm.getSubcategoryTitle(post.SubcategoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get subcategory title: %w", err)
		}
		post.Subcategory = Subcategory{ID: post.SubcategoryID, Title: title}

		// Fetch comments for the post
		comments, err := pm.GetComments(post.ID)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to retrieve comments: %v", err)
			return nil, fmt.Errorf("failed to retrieve comments: %w", err)
		}
		post.Comments = comments

		posts = append(posts, post)
	}

	return posts, nil
}

func (pm *PostModel) GetComments(postID int) ([]Comment, error) {
	// Execute the query to retrieve comments
	rows, err := pm.DB.Query(dbserver.CommentQuery, postID)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to retrieve comments: %v", err)
		return nil, fmt.Errorf("failed to retrieve comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment

	// Iterate over the rows returned by the query
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.AuthorName, &c.AuthorImage, &c.Content, &c.CreatedAt, &c.Likes, &c.Dislikes); err != nil {
			logger.ErrorLogger.Printf("Failed to scan comments: %v", err)
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		// Get replies associated with the comment
		c.Replies, err = pm.GetReplies(c.ID)
		if err != nil {
			logger.ErrorLogger.Printf("failed to retrieve replies for comment: %v", err)
			return nil, fmt.Errorf("failed to retrieve replies for comment %d: %w", c.ID, err)
		}
		comments = append(comments, c)
	}

	return comments, nil
}

func (pm *PostModel) GetReplies(commentID int) ([]Reply, error) {
	// Execute the query to retrieve replies
	rows, err := pm.DB.Query(dbserver.ReplyQuery, commentID)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to retrieve replies: %v", err)
		return nil, fmt.Errorf("failed to retrieve replies: %w", err)
	}
	defer rows.Close()

	var replies []Reply
	// Iterate over the rows returned by the query
	for rows.Next() {
		var reply Reply
		if err := rows.Scan(&reply.ID, &reply.CommentID, &reply.AuthorID, &reply.AuthorName, &reply.Content, &reply.AuthorImage, &reply.CreatedAt); err != nil {
			logger.ErrorLogger.Printf("Failed to scan reply: %v", err)
			return nil, fmt.Errorf("failed to scan reply: %w", err)
		}
		replies = append(replies, reply)
	}

	return replies, nil
}

// getSubcategoryTitle retrieves the title of a subcategory by its ID
func (pm *PostModel) getSubcategoryTitle(subcategoryID int) (string, error) {
	var title string
	query := "SELECT title FROM subcategories WHERE id = ?"
	err := pm.DB.QueryRow(query, subcategoryID).Scan(&title)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		logger.ErrorLogger.Printf("Failed to retrieve subcategory title: %v", err)
		return "", fmt.Errorf("failed to retrieve subcategory title: %w", err)
	}
	return title, nil
}

// GetUserPostReactions retrieves the like/dislike status of a user for a specific post
func (pm *PostModel) GetUserPostReactions(post *Post, userID int) error {
	reactionStmt := `
        SELECT
            EXISTS (SELECT 1 FROM likes WHERE post_id = ? AND user_id = ?) AS has_liked,
            EXISTS (SELECT 1 FROM dislikes WHERE post_id = ? AND user_id = ?) AS has_disliked
    `

	err := pm.DB.QueryRow(reactionStmt, post.ID, userID, post.ID, userID).Scan(&post.HasLiked, &post.HasDisliked)
	if err != nil {
		fmt.Println(err)
		logger.ErrorLogger.Printf("Failed to query reactions: %v", err)
		return fmt.Errorf("failed to query reactions: %w", err)
	}

	return nil
}

// GetUserCommentReactions retrieves the like/dislike status of a user for a specific comment
func (pm *PostModel) GetUserCommentReactions(comment *Comment, userID int) error {
	reactionStmt := `
        SELECT
            EXISTS (SELECT 1 FROM likes WHERE comment_id = ? AND user_id = ?) AS has_liked,
            EXISTS (SELECT 1 FROM dislikes WHERE comment_id = ? AND user_id = ?) AS has_disliked
    `

	err := pm.DB.QueryRow(reactionStmt, comment.ID, userID, comment.ID, userID).Scan(&comment.HasLiked, &comment.HasDisliked)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to query reactions: %v", err)
		return fmt.Errorf("failed to query reactions: %w", err)
	}

	return nil
}

func (pm *PostModel) CreateComment(postID int, authorID int, content string, authorName string) (int, error) {
	sanitizedContent := utils.ParseAndReplaceImageTags(content)

	// SQL statement to insert a new comment
	stmt := `INSERT INTO comments (post_id, author_id, content, author_name) VALUES (?, ?, ?, ?)`
	result, err := pm.DB.Exec(stmt, postID, authorID, sanitizedContent, authorName)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to insert comment: %v", err)
		return 0, fmt.Errorf("failed to insert comment: %w", err)
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	logger.InfoLogger.Printf("Comment inserted by user with ID %d", authorID)
	return int(commentID), nil
}

func (pm *PostModel) CreateReply(commentID int, authorID int, content string, authorName string) (int, error) {
	sanitizedContent := utils.ParseAndReplaceImageTags(content)

	// SQL statement to insert a new reply
	stmt := `INSERT INTO replies (comment_id, author_id, content, author_name) VALUES (?, ?, ?, ?)`
	result, err := pm.DB.Exec(stmt, commentID, authorID, sanitizedContent, authorName)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to insert reply: %v", err)
		return 0, fmt.Errorf("failed to insert reply: %w", err)
	}

	replyID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	logger.InfoLogger.Printf("Reply inserted by user with ID %d", authorID)
	return int(replyID), nil
}

func (pm *PostModel) SubmitComment(r *http.Request, postModel *PostModel, postID int, userID int, username string) error {
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		logger.ErrorLogger.Printf("Error parsing form: %v", err)
		return fmt.Errorf("error parsing form: %w", err)
	}
	comment := r.FormValue("comment")

	if comment != "" {
		_, err := postModel.CreateComment(postID, userID, comment, username)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to create comment: %v", err)
			return fmt.Errorf("failed to create comment: %w", err)
		}
	}

	return nil
}

func (pm *PostModel) SubmitReply(r *http.Request, postModel *PostModel, commentID int, userID int, username string) error {
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		logger.ErrorLogger.Printf("Error parsing form: %v", err)
		return fmt.Errorf("error parsing form: %w", err)
	}
	comment := r.FormValue("comment")

	if comment != "" {
		_, err := postModel.CreateReply(commentID, userID, comment, username)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to create reply: %v", err)
			return fmt.Errorf("failed to create reply: %w", err)
		}
	}

	return nil
}

func (pm *PostModel) InsertReaction(table string, column string, id int, userID int, isLike bool) error {
	var oppositeReactionTable string
	if table == "likes" {
		oppositeReactionTable = "dislikes"
	} else {
		oppositeReactionTable = "likes"
	}
	// Begin a new transaction
	tx, err := pm.DB.Begin()
	if err != nil {
		logger.ErrorLogger.Printf("Failed to begin transaction: %v", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	var existingReaction bool
	err = tx.QueryRow("SELECT EXISTS (SELECT 1 FROM "+table+" WHERE "+column+" = ? AND user_id = ?)", id, userID).Scan(&existingReaction)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to check for existing reaction: %v", err)
		return fmt.Errorf("failed to check for existing reaction: %w", err)
	}

	// Delete existing reaction if it exists
	if existingReaction {
		_, err = tx.Exec("DELETE FROM "+table+" WHERE "+column+" = ? AND user_id = ?", id, userID)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to delete existing reaction: %v", err)
			return fmt.Errorf("failed to delete existing reaction: %w", err)
		}
	} else {
		// Insert new reaction
		_, err = tx.Exec("INSERT INTO "+table+" ("+column+", user_id) VALUES (?, ?)", id, userID)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to insert new reaction %v", err)
			return fmt.Errorf("failed to insert new reaction: %w", err)
		}
		// Delete the opposite reaction if it exists
		_, err = tx.Exec("DELETE FROM "+oppositeReactionTable+" WHERE "+column+" = ? AND user_id = ?", id, userID)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to delete the opposite reaction: %v", err)
			return fmt.Errorf("failed to delete the opposite reaction: %w", err)
		}
	}
// Commit the transaction
err = tx.Commit()
if err != nil {
    logger.ErrorLogger.Printf("Failed to commit transaction: %v", err)
    return fmt.Errorf("failed to commit transaction: %w", err)
}

	return nil
}
