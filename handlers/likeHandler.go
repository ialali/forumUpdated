package handlers

import (
	"database/sql"
	"forum/database"
	"net/http"
	"strconv"
)

// Separate Go handler for liking a post
func LikePostHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		// Get the authenticated user data
		userData := GetAuthenticatedUserData(db, r)
		// If the user is not authenticated, redirect to the login page
		if !userData.IsAuthenticated {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		postID, err := strconv.Atoi(r.FormValue("post_id"))
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}
		// Retrieve the user ID from the session
		userID, ok := GetAuthenticatedUserID(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		err = database.InsertPostLike(db, userID, postID)
		if err != nil {
			http.Error(w, "error inserting the like into database", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.Error(w, "error handling likes ", http.StatusSeeOther)

}
