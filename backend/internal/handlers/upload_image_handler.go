package handlers

import (
	"database/sql"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"literary-lions/internal/models"
	"literary-lions/pkg/logger"
)

func UploadImageHandler(w http.ResponseWriter, r *http.Request, sessionModel *models.SessionModel, db *sql.DB) {

	// Ensure the request method is POST
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form with a max memory of 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing multipart form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve the uploaded file
	file, header, err := r.FormFile("profile-img")
	if err != nil {
		http.Error(w, "Error retrieving the file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Failed to decode image: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Resize the image to 768x768 pixels
	resizedImg := ResizeImage(img, 768, 768)

	// Create the directory for storing the image if it does not exist
	directory := "/app/frontend/static/uploads"
	if err := os.MkdirAll(directory, 0755); err != nil {
		http.Error(w, "Failed to create directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve the user ID from the session
	userID, err := sessionModel.GetUserID(r)
	if err != nil {
		http.Error(w, "Failed to retrieve user ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine the file extension and set the file path
	ext := strings.ToLower(filepath.Ext(header.Filename))
	fileName := fmt.Sprintf("%d.jpg", userID)
	filePath := filepath.Join(directory, fileName)

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Unable to create the file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Encode and save the resized image based on its file extension
	switch ext {
	case ".jpeg", ".jpg":
		err = jpeg.Encode(dst, resizedImg, nil)
	case ".png":
		err = png.Encode(dst, resizedImg)
	default:
		http.Error(w, "Unsupported file type: "+ext, http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Failed to encode and save resized image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the user's profile picture path in the database
	if _, err = db.Exec("UPDATE users SET profile_picture_path = ? WHERE id = ?", fileName, userID); err != nil {
		http.Error(w, "Database update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.InfoLogger.Printf("Successfully Uploaded and Resized Profile Picture for user ID %d", userID)
	http.Redirect(w, r, fmt.Sprintf("/profile/%d", userID), http.StatusSeeOther)
}

func ResizeImage(src image.Image, newWidth, newHeight int) image.Image {
	// Create a new RGBA image with the specified dimensions
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Calculate the scaling ratios
	xRatio := float64(src.Bounds().Dx()) / float64(newWidth)
	yRatio := float64(src.Bounds().Dy()) / float64(newHeight)

	// Perform the resizing
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) * xRatio)
			srcY := int(float64(y) * yRatio)
			srcColor := src.At(srcX, srcY)
			dst.Set(x, y, srcColor)
		}
	}
	return dst
}
