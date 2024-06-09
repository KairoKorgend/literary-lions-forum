package api

import (
	"database/sql"
	"io/ioutil"
	"literary-lions/internal/models"
	"literary-lions/pkg/logger"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Discard logger output during tests
	logger.InfoLogger.SetOutput(ioutil.Discard)
	logger.ErrorLogger.SetOutput(ioutil.Discard)

	// Run the tests
	os.Exit(m.Run())
}

func TestRegisterRoutes(t *testing.T) {
	mux := http.NewServeMux()
	sessionModel := &models.SessionModel{}
	userModel := &models.UserModel{}
	postModel := &models.PostModel{}
	categoryModel := &models.CategoryModel{}
	db := &sql.DB{}

	err := RegisterRoutes(mux, sessionModel, userModel, postModel, categoryModel, db)
	if err != nil {
		t.Fatalf("Failed to register routes: %v", err)
	}

	// Test the "/" route
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test the "/login" route
	req, err = http.NewRequest("GET", "/login", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
