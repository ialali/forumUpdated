package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"forum/database"
	auth "forum/middleware"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
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
func GoogleRegisterHandler(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func GithubRegisterHandler(w http.ResponseWriter, r *http.Request) {
	url := githubOauthConfig.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusFound)
}

func GoogleCallbackHandler2(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get user info
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Generate a hashed password
	password := "randomlyGeneratedPassword" // Replace with actual random password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get username from email
	username := strings.Split(userInfo.Email, "@")[0]

	// Get current time
	registrationDate := time.Now()

	// Insert user into database
	result, err := db.Exec("INSERT INTO users (username, email, password, registration_date) VALUES (?, ?, ?, ?)", username, userInfo.Email, string(hashedPassword), registrationDate)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	userID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

func GithubCallbackHandler2(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	token, err := githubOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get user info
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Generate a hashed password
	password := "randomlyGeneratedPassword" // Replace with actual random password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get username from email
	username := strings.Split(userInfo.Email, "@")[0]

	// Get current time
	registrationDate := time.Now()

	// Insert user into database
	result, err := db.Exec("INSERT INTO users (username, email, password, registration_date) VALUES (?, ?, ?, ?)", username, userInfo.Email, string(hashedPassword), registrationDate)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	userID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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