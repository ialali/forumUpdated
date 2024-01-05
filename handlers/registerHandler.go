package handlers

import (
	"database/sql"
	"fmt"
	"forum/database"
	auth "forum/middleware"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/google/uuid"
)

func RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	tmpl.Execute(w, nil)
}

func RegisterSubmitHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return // Return to exit the function
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return // Return to exit the function
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	if username == "" || password == "" {
		http.Error(w, "Please fill in username or password", http.StatusBadRequest)
		return // Return to exit the function
	}

	if auth.IsDuplicateUser(db, username, email) {
		http.Error(w, "Username or email is already taken", http.StatusConflict)
		return
	}

	userID, err := database.RegisterUser(db, username, email, password)
	if err != nil {
		http.Error(w, "Registration Failure", http.StatusInternalServerError)
		return // Return to exit the function
	}

	sessionToken := uuid.New().String()

	mu.Lock()
	defer mu.Unlock()

	intUserID := int(userID)

	// Check if the user already has an active session
	if existingSessionID, ok := userSessions[intUserID]; ok {
		// If so, remove the existing session
		delete(userSessions, intUserID)
		log.Printf("Removed existing session for user %d\n", userID)
		// Also delete the session from the sessions map
		delete(sessions, existingSessionID)
	}

	// Store the session ID and user ID in their respective maps
	userSessions[intUserID] = sessionToken
	sessions[sessionToken] = intUserID
	log.Println(sessions)
	log.Println(sessionToken)

	// Store the session ID in a cookie with an expiration time
	expiration := time.Now().Add(24 * time.Hour) // 24 hours
	cookie := http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
		Secure:   true, // Enable only in production with HTTPS
	}

	http.SetCookie(w, &cookie)

	// Redirect the user to the home page after successful registration
	http.Redirect(w, r, "/", http.StatusSeeOther)
	fmt.Println(userID)
}
