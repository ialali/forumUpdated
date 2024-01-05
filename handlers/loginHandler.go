package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"forumUpdated/database"
	auth "forumUpdated/middleware"
	"log"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type UserInfo struct {
	Email string `json:"email"`
	// Include other fields as needed
}
type GithubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:1219/auth/google/callback",
		ClientID:     "client_id", // Replace with your client ID
		ClientSecret: "client_secret", // Replace with your client secret
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	githubOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:1219/auth/github/callback",
		ClientID:     "client_id", // Replace with your client ID
		ClientSecret: "client_secret", // Replace with your client secret
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
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
func HsandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
func HandleGoogleCallback(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	client := googleOauthConfig.Client(r.Context(), token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	// Parse the user info from the response
	userInfo, err := parseUserInfo(response)
	if err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	// Check the database for a user with the retrieved email address
	user, err := database.GetUserByEmail(db, userInfo.Email)
	if err != nil {
		user, err = database.CreateUser(db, userInfo.Email)
        if err != nil {
			log.Println(err)
            http.Error(w, "Failed to create user", http.StatusInternalServerError)
            return
        }
	}
	fmt.Println(user)

	userID := user.ID
	fmt.Println(userID)

	// Create a new session for the user
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
}
func parseUserInfo(response *http.Response) (UserInfo, error) {
	var userInfo UserInfo
	err := json.NewDecoder(response.Body).Decode(&userInfo)
	if err != nil {
		return UserInfo{}, err
	}
	return userInfo, nil
}
func HandleGithubLogin(w http.ResponseWriter, r *http.Request) {
	url := githubOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
func HandleGithubCallback(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	token, err := githubOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	client := githubOauthConfig.Client(r.Context(), token)
	response, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	// Parse the user info from the response
	userInfo, err := parseGithubUserInfo(response)
	if err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	// Check the database for a user with the retrieved email address
	user, err := database.GetUserByEmail(db, userInfo.Email)
	if err != nil {
		http.Error(w, "Failed to get user by email", http.StatusInternalServerError)
		return
	}
	fmt.Println(user)

	userID := user.ID
	fmt.Println(userID)

	// Create a new session for the user
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
}
func parseGithubUserInfo(response *http.Response) (GithubEmail, error) {
	var emails []GithubEmail
	err := json.NewDecoder(response.Body).Decode(&emails)
	if err != nil {
		return GithubEmail{}, err
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email, nil
		}
	}

	return GithubEmail{}, errors.New("no primary, verified email found")
}
