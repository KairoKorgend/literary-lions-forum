package models

import (
	"database/sql"
	"fmt"
	"literary-lions/pkg/logger"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID
	UserID    int
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionModel struct {
	DB *sql.DB
}

func NewSessionModel() (*SessionModel, error) {
	db, err := sql.Open("sqlite3", DatabaseLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &SessionModel{DB: db}, nil
}

func (m *SessionModel) GetUserID(r *http.Request) (int, error) {
	// Get the session ID from the cookie
	cookie, err := r.Cookie("session")
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get session cookie: %v\n", err)
		return 0, fmt.Errorf("failed to get session cookie: %v", err)
	}

	// Parse the session ID from the cookie value
	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to parse session ID: %v\n", err)
		return 0, fmt.Errorf("failed to parse session ID: %v", err)
	}

	// Retrieve the session from the database using the session ID
	session, err := m.Get(sessionID)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get session: %v\n", err)
		return 0, fmt.Errorf("failed to get session: %v", err)
	}

	// Check if the session was found
	if session == nil {
		return 0, fmt.Errorf("no session found with ID '%s'", sessionID)
	}

	logger.InfoLogger.Printf("Retrieved session: %+v\n", session)

	// REturn the user ID from the session
	return session.UserID, nil
}

func (m *SessionModel) Create(w http.ResponseWriter, userID int, sessionID uuid.UUID) error {
	// SQL statement to insert a new session into the database
	stmt := `INSERT INTO sessions (id, user_id, created_at, expires_at)
	VALUES (?, ?, datetime('now'), datetime('now', '+1 hour'))`

	// Execute the SQL statement
	_, err := m.DB.Exec(stmt, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to insert a new session: %w", err)
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID.String(),
		Path:     "/",
		HttpOnly: true,
	})

	return nil
}

func (m *SessionModel) Get(sessionID uuid.UUID) (*Session, error) {
	// SQL statement to retrieve a session from the database
	stmt := "SELECT id, user_id, created_at, expires_at FROM sessions WHERE id = ?"

	// Execute the SQL statement and scan the result into a Session struct
	row := m.DB.QueryRow(stmt, sessionID)
	s := &Session{}

	err := row.Scan(&s.ID, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// No matching session found
			return nil, nil
		}
		return nil, fmt.Errorf("failed row scan: %w", err)
	}

	return s, nil
}

func (m *SessionModel) Delete(sessionID uuid.UUID) error {
	// SQL statement to delete a session from the database
	stmt := "DELETE FROM sessions WHERE id = ?"

	// Execute the SQL statement
	_, err := m.DB.Exec(stmt, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (m *SessionModel) Exists(sessionID uuid.UUID) bool {
	// Retrieve the session from the database
	s, err := m.Get(sessionID)
	if err != nil || s == nil || time.Now().After(s.ExpiresAt) {
		// If there's an error, the session doesn't exist or has expired
		return false
	}

	// The session exists and has not expired
	return true
}

func (m *SessionModel) IsAuthenticated(r *http.Request) (bool, int, error) {
	// Get the session ID from the cookie
	cookie, err := r.Cookie("session")
	if err != nil {
		// If there's an error, the cookie doesn't exist, so the user is not authenticated
		return false, 0, nil
	}

	// Parse the session ID from the cookie
	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		// If there's an error, the session ID is invalid, so the user is not authenticated
		return false, 0, nil
	}

	// Check whether the session exists
	session, err := m.Get(sessionID)
	if err != nil || session == nil {
		// If there's an error or the session doesn't exist, the user is not authenticated
		return false, 0, err
	}

	// The session exists and has not expired, so the user is authenticated
	return true, session.UserID, nil
}
