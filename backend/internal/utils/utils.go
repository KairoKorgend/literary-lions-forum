package utils

import (
	"encoding/base64"
	"fmt"
	dbserver "literary-lions/internal/db"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

// ParseURL parses the URL from the request to extract the post ID, action, and comment ID if applicable
func ParseURL(r *http.Request) (int, bool, string, int, error) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		return 0, false, "", 0, fmt.Errorf("invalid URL path")
	}

	// Extract and convert the post ID from the URL path
	postID, err := strconv.Atoi(pathParts[2])
	if err != nil {
		return 0, false, "", 0, fmt.Errorf("invalid post ID: %w", err)
	}

	showModal := false
	action := ""
	commentID := 0

	// Check if the URL path contains additional action information
	if len(pathParts) >= 4 {
		switch pathParts[3] {
		case "comment":
			showModal = true
			action = "comment"
		case "reply":
			showModal = true
			action = "reply"
			// Extract the comment ID
			if len(pathParts) >= 5 {
				commentID, err = strconv.Atoi(pathParts[4])
				if err != nil {
					return postID, showModal, action, 0, fmt.Errorf("invalid comment ID: %w", err)
				}
			}
		case "delete":
			action = "delete"
		}
	}

	return postID, showModal, action, commentID, nil
}

// ConvertToDataURL converts image binary data to a Base64-encoded data URL.
func ConvertToDataURL(image []byte) string {
	if len(image) == 0 {
		return ""
	}
	mimeType := http.DetectContentType(image)
	encoded := base64.StdEncoding.EncodeToString(image)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
}

// TimeAgo returns a human-readable string representing the time elapsed since the given time
func TimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
	case diff < time.Hour*24:
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	default:
		return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
	}
}

// ParseAndReplaceImageTags converts the image url inside of the tags into a visible and clickable image
func ParseAndReplaceImageTags(content string) string {
	imgTagRegex := regexp.MustCompile(`\[img\](.*?)\[/img\]`)
	sanitizedContent := bluemonday.UGCPolicy().Sanitize(content)
	matches := imgTagRegex.FindAllStringSubmatchIndex(sanitizedContent, -1)
	replacedContent := sanitizedContent

	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		if len(match) >= 4 {
			urlStart := match[2]
			urlEnd := match[3]

			url := sanitizedContent[urlStart:urlEnd]

			imgTag := fmt.Sprintf(`<a href="%s" data-lightbox="post-images" class="post__img" data-title="Image" data-type="image"><img src="%s" alt="Post Image"></a>`, url, url)

			replacedContent = replacedContent[:match[0]] + imgTag + replacedContent[match[1]:]
		}
	}

	return replacedContent
}

func IndexGetPostsQueryBuilder(subcategoryId int, searchString string, filter string) (string, []interface{}) {
	var queryString string
	var orderBy string
	var sortOrder string
	var queryParams []interface{}

	// Determine the order by and sort order based on the filter
	switch filter {
	case "newest":
		orderBy = "p.event_time"
		sortOrder = "DESC"
	case "oldest":
		orderBy = "p.event_time"
		sortOrder = "ASC"
	case "most_comments":
		orderBy = "comment_count"
		sortOrder = "DESC"
	case "most_likes":
		orderBy = "like_count"
		sortOrder = "DESC"
	case "most_dislikes":
		orderBy = "dislike_count"
		sortOrder = "DESC"
	default:
		orderBy = "p.event_time"
		sortOrder = "DESC"
	}

	// Build the query string based on the subcategory ID and search string
	if subcategoryId == 0 && searchString == "" {
		queryString = dbserver.IndexPostQuery
	} else if subcategoryId == 0 && searchString != "" {
		queryString = dbserver.SearchPostQuery
		queryParams = append(queryParams, "%"+searchString+"%")
	} else {
		queryString = dbserver.CategoryPostQuery
		queryParams = append(queryParams, strconv.Itoa(subcategoryId))
	}

	queryString = fmt.Sprintf(queryString, orderBy, sortOrder)

	return queryString, queryParams
}

func InitTmpl(tmplName string) (*template.Template, error) {
	var tmpl *template.Template

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve current directory path: %w", err)
	}

	// Build the template path
	tmplPath := filepath.Join(cwd, "..", "..", "frontend", "templates", tmplName)

	// Parse the template file
	tmpl, err = template.ParseFiles(tmplPath)
	if err != nil {
		log.Fatalf("error parsing template: %v", err)
	}

	return tmpl, nil
}
