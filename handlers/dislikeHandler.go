package handlers

import (
	"database/sql"
	"forum/database"
	"net/http"
	"strconv"
)

func DislikePostHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method.", http.StatusBadRequest)
		return

	}
	userData := GetAuthenticatedUserData(db, r)
	if !userData.IsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	userID, ok := GetAuthenticatedUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	err = database.InsertPostDislike(db, userID, postID)
	if err != nil {
		http.Error(w, "Error inserting dislike into the database", http.StatusSeeOther)
		return

	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
