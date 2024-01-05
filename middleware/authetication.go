package auth

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func IsDuplicateUser(db *sql.DB, username, email string) bool {
	// Perform a database query to check if the username or email already exists in the database.
	// Return true if duplicate, false if not.
	var count int
	query := "SELECT COUNT(*) FROM users WHERE username = ? OR email = ?"
	row := db.QueryRow(query, username, email)
	if err := row.Scan(&count); err != nil {
		// Handle the error, e.g., log it or return false.
		return false
	}

	return count > 0
}
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
func GetHashedPassword(db *sql.DB, username string) (string, error) {
	// Query the database to get the hashed password for the provided username.
	var hashedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		// Handle errors, e.g., username not found in the database.
		if err == sql.ErrNoRows {
			return "", errors.New("User not found")
		}
		return "", err
	}

	return hashedPassword, nil
}
