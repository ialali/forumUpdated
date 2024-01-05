package handlers

import "net/http"

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		// Handle the error, log it, or redirect to an error page
		http.Error(w, "Failed to get session cookie", http.StatusInternalServerError)
		return
	}

	if cookie != nil {
		delete(sessions, cookie.Value)
		ClearSession(w, cookie.Value)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ClearSession(w http.ResponseWriter, sessionName string) {
	cookie := http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, &cookie)

	// Remove the session data from the map
	delete(sessions, sessionName)
}
