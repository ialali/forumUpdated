package main

import (
	"fmt"
	"forumUpdated/database"
	"forumUpdated/handlers"

	"log"
	"net/http"
)

func main() {
	// Define the path to your SQLite database file
	dbPath := "database/database.db"

	// Open a connection to the database
	db, err := database.OpenDatabase(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize the schema and create tables
	err = database.InitializeSchema(db)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.ShowPostHandler(w, r, db)
	})

	http.HandleFunc("/register", handlers.RegisterPageHandler)
	http.HandleFunc("/registerauth", func(w http.ResponseWriter, r *http.Request) {
		handlers.RegisterSubmitHandler(w, r, db)
	})
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/loginauth", func(w http.ResponseWriter, r *http.Request) {
		handlers.LoginSubmitHandler(w, r, db)
	})
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/create-post", handlers.AddPost)
	http.HandleFunc("/add-post", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddPostSubmit(w, r, db)
	})
	http.HandleFunc("/add-comment", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddCommentHandler(w, r, db)
	})
	http.HandleFunc("/like-post", func(w http.ResponseWriter, r *http.Request) {
		handlers.LikePostHandler(w, r, db)
	})
	http.HandleFunc("/dislike-post", func(w http.ResponseWriter, r *http.Request) {
		handlers.DislikePostHandler(w, r, db)
	})
	http.HandleFunc("/like-comment", func(w http.ResponseWriter, r *http.Request) {
		handlers.LikeCommentHandler(w, r, db)
	})
	http.HandleFunc("/dislike-comment", func(w http.ResponseWriter, r *http.Request) {
		handlers.DisikeCommentHandler(w, r, db)
	})
	http.HandleFunc("/filter", func(w http.ResponseWriter, r *http.Request) {
		handlers.FilterPosts(w, r, db)
	})
	http.HandleFunc("/auth/google/login", handlers.HsandleGoogleLogin)
	http.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleGoogleCallback(db, w, r)
	})
	http.HandleFunc("/auth/github/login", handlers.HandleGithubLogin)
	http.HandleFunc("/auth/github/callback", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleGithubCallback(db, w, r)
	})
	http.HandleFunc("/auth/google/register", handlers.GoogleRegisterHandler)
	http.HandleFunc("/auth/google/registerauth", func(w http.ResponseWriter, r *http.Request) {
		handlers.GoogleCallbackHandler2(db, w, r)
	})
	http.HandleFunc("/auth/github/register", handlers.GithubRegisterHandler)
	http.HandleFunc("/auth/github/registerauth", func(w http.ResponseWriter, r *http.Request) {
		handlers.GithubCallbackHandler2(db, w, r)
	})

	fmt.Println("server started on http://localhost:1219")
	http.ListenAndServe(":1219", nil)

	// You can now use the 'db' connection to perform database operations.

}
