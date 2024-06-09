package models

import (
	"database/sql"
	"errors"
	"fmt"
	"literary-lions/pkg/logger"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const DatabaseLocation = "../internal/db/literary_lions.db"

type User struct {
	ID                 int
	Login              string
	Email              string
	HashedPassword     []byte
	Created            time.Time
	CreatedAtFormatted string
	ProfilePicturePath string
}

type UserModel struct {
	DB *sql.DB
}

func NewUserModel() (*UserModel, error) {
	db, err := sql.Open("sqlite3", DatabaseLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &UserModel{DB: db}, nil
}

func (m *UserModel) Get(id int) (*User, error) {
	// SQL statement to retrieve a user from the database
	stmt := "SELECT id, login, email, created, profile_picture_path FROM users WHERE id = ?"

	// Execute the SQL statement and scan the result into a User struct
	row := m.DB.QueryRow(stmt, id)
	user := &User{}

	err := row.Scan(&user.ID, &user.Login, &user.Email, &user.Created, &user.ProfilePicturePath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord // No matching user found
		} else {
			return nil, err
		}
	}

	// Format the creation date
	user.CreatedAtFormatted = user.Created.Format("January 2, 2006")

	return user, nil
}

func (m *UserModel) Insert(name, email, password string) error {
	// Hash the user's password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		logger.ErrorLogger.Printf("Error hashing password: %v", err)
		return fmt.Errorf("error hashing password: %w", err)
	}

	// SQL statement to insert a new user
	stmt := `INSERT INTO users (email, login, hashed_password, created)
	VALUES(?, ?, ?, datetime('now'))`

	result, err := m.DB.Exec(stmt, email, name, string(hashedPassword))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			return ErrDuplicateEmail // Email already exists
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.login") {
			return ErrDuplicateLogin // Login name already exists
		}
		return err
	}
	// Get the ID of the newly inserted user
	userID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting the id of newly inserted user: %w", err)
	}
	// Set the default profile picture for the new user
	if err := setDefaultProfilePicture(m.DB, userID); err != nil {
		log.Printf("Failed to set default profile picture: %v", err)
	}

	return nil
}

// Sets the default profile picture for a new user
func setDefaultProfilePicture(db *sql.DB, userID int64) error {
	_, err := db.Exec("UPDATE users SET profile_picture_path = ? WHERE id = ?", "default-img.jpg", userID)
	return fmt.Errorf("error setting default image for user: %w", err)
}

func (m *UserModel) Authenticate(email, password string) (*User, error) {
	user := &User{}
	// SQL statement to retrieve the user's hashed password
	stmt := "SELECT id, email, hashed_password FROM users WHERE email = ?"

	err := m.DB.QueryRow(stmt, email).Scan(&user.ID, &user.Email, &user.HashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials // Invalid credentials
		} else {
			return nil, err
		}
	}

	// Compare the hashed password with the provided password
	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials // Invalid credentials
	}

	return user, nil
}

func (m *UserModel) PasswordUpdate(id int, currentPassword, newPassword string) error {
	var currentHashedPassword []byte
	// SQL statement to retrieve the current hashed password
	stmt := "SELECT hashed_password FROM users WHERE id = ?"
	err := m.DB.QueryRow(stmt, id).Scan(&currentHashedPassword)
	if err != nil {
		return fmt.Errorf("error retrieving current hashed password: %w", err)
	}
	// Compare the current hashed password with the provided current password
	err = bcrypt.CompareHashAndPassword(currentHashedPassword, []byte(currentPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials // Invalid current password
		} else {
			return err
		}
	}
	// Hash the new password
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}
	// SQL statement to update the user's password
	stmt = "UPDATE users SET hashed_password = ? WHERE id = ?"
	_, err = m.DB.Exec(stmt, string(newHashedPassword), id)
	return fmt.Errorf("error updating user's password: %w", err)
}

func (m *UserModel) PostCount(id int) (int, error) {
	// SQL statement to count the number of posts by the user
	stmt := "SELECT COUNT(*) FROM posts WHERE author_id = ?"

	var count int
	err := m.DB.QueryRow(stmt, id).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting the number of posts by the user: %w", err)
	}

	return count, nil
}

func (m *UserModel) GetUserCreationDateByAuthorID(authorID int) (string, error) {
	// SQL statement to retrieve the user's creation date
	stmt := `SELECT u.created 
             FROM users u 
             INNER JOIN posts p ON u.id = p.author_id 
             WHERE p.author_id = ?`

	var createdAt time.Time
	err := m.DB.QueryRow(stmt, authorID).Scan(&createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNoRecord // No matching record found
		} else {
			return "", err
		}
	}

	// Format the Created time to a user-friendly string
	CreatedAtFormatted := createdAt.Format("January 2, 2006")

	return CreatedAtFormatted, nil
}
