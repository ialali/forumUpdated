package handlers

import (
	"database/sql"
	"forum/database"
	"log"
	"net/http"
	"text/template"
)

func AddPost(w http.ResponseWriter, r *http.Request) {
	_, isAuthenticated := GetAuthenticatedUserID(r)
	if !isAuthenticated {
		// Redirect to the login page or show an error message.
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return // Return to avoid further execution if not authenticated
	}

	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		http.Error(w, "error parsing the template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)

}

func AddPostSubmit(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != "POST" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Check if the user is authenticated
	userID, isAuthenticated := GetAuthenticatedUserID(r)
	if !isAuthenticated {
		// Redirect to the login page or show an error message.
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	allowedCategories := []string{"Sport", "Nutrition", "Recovery", "Tech", "Other"}
	selectedCategory := r.FormValue("category")

	if !contains(allowedCategories, selectedCategory) {
		http.Error(w, "Invalid category", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		http.Error(w, "Missing required fields", http.StatusUnprocessableEntity)
		return
	}

	err := database.InsertPost(db, selectedCategory, title, content, userID)
	if err != nil {
		log.Println("Error inserting post:", err)
		http.Redirect(w, r, "/error/500", http.StatusSeeOther)
		return
	}

	// Redirect the user to the home page after successfully adding the post
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func contains(arr []string, item string) bool {
	for _, el := range arr {
		if el == item {
			return true
		}
	}
	return false
}
