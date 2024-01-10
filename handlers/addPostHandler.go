package handlers

import (
	"database/sql"
	"forumUpdated/database"
	"io"
	"log"
	"net/http"
	"os"
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

	// Check the size of the uploaded file
	err := r.ParseMultipartForm(20 << 20) // limit your max input length to 20MB
	if err != nil {
		http.Error(w, "The uploaded file is too big. Please choose an image that's less than 20MB in size.", http.StatusBadRequest)
		return
	}

	var imagePath sql.NullString
	file, header, err := r.FormFile("image") // retrieve the file from form data

	// Check the size of the uploaded file
	if err != nil {
		if err == http.ErrMissingFile {
			imagePath.Valid = false
		} else {
			http.Error(w, "Error retrieving the file", http.StatusInternalServerError)
			return
		}
	} else {
		// Check the size of the uploaded file
		if header.Size > 20<<20 {
			http.Error(w, "The uploaded file is too big. Please choose an image that's less than 20MB in size.", http.StatusBadRequest)
			return
		}

		defer file.Close()

		// Create a new file in the uploads directory
		dst, err := os.Create("uploads/" + header.Filename)
		if err != nil {
			log.Println("Error creating file:", err)
			http.Error(w, "Error creating file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			log.Println("Error copying file:", err)
			http.Error(w, "Error copying file", http.StatusInternalServerError)
			return
		}

		// Use the path of the new file as the image path
		imagePath.String = "./uploads/" + header.Filename
		imagePath.Valid = true
	}

	if title == "" || content == "" {
		http.Error(w, "Missing required fields", http.StatusUnprocessableEntity)
		return
	}

	err = database.InsertPost(db, selectedCategory, title, content, imagePath, userID)
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
