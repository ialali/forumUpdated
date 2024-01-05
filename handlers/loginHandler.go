package handlers

import (
	"database/sql"
	"forum/database"
	auth "forum/middleware"
	"log"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	sessions     = make(map[string]int)
	userSessions = make(map[int]string)
	mu           sync.Mutex
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, "Error rendering login page", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)

}

func LoginSubmitHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		http.Error(w, "Please fill in username or password", http.StatusBadRequest)
		return // Return to exit the function
	}
	if !database.UserExists(db, username) {
		// Redirect to the login page with an appropriate message
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	// Retrieve the hashed password from the database for the given username.
	storedHashedPassword, err := auth.GetHashedPassword(db, username)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Use bcrypt.CompareHashAndPassword to check if the provided password matches the stored hashed password.
	err = bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(password))
	if err != nil {
		http.Error(w, "Incorrect username or password", http.StatusUnauthorized)
		return
	}
	userID, err := database.GetIDBYusername(db, username)
	if err != nil {
		log.Println("Error getting user ID:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("User ID:", userID)

	// Generate a new UUID for the session
	sessionToken := uuid.New().String()

	mu.Lock()
	defer mu.Unlock()

	// Check if the user already has an active session
	if existingSessionID, ok := userSessions[userID]; ok {
		// If so, remove the existing session
		delete(userSessions, userID)
		log.Printf("Removed existing session for user %d\n", userID)
		// Also delete the session from the sessions map
		delete(sessions, existingSessionID)
	}

	// Store the session ID and user ID in their respective maps
	userSessions[userID] = sessionToken
	sessions[sessionToken] = userID
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

	// Redirect the user to the home page after successful login
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}
func IsAuthenticated(r *http.Request) bool {
	// Check if the user is authenticated by looking for a session token.
	cookie, err := r.Cookie("session_token")
	if err != nil {
		// No session token found, the user is not authenticated.
		log.Println("No session token found.")
		return false
	}

	// Retrieve the session token from the cookie.
	sessionToken := cookie.Value
	log.Println("Retrieved session token:", sessionToken)

	// Look up the user's ID associated with the session token.
	mu.Lock()
	defer mu.Unlock()
	_, ok := sessions[sessionToken]

	if ok {
		log.Println("User is authenticated.")
	} else {
		log.Println("User is not authenticated.")
	}

	// If the session token is found in the sessions map, the user is authenticated.
	return ok
}
func GetAuthenticatedUserID(r *http.Request) (int, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return 0, false
	}
	userID, ok := sessions[cookie.Value]
	return userID, ok
}
