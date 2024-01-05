package main

import (
	"fmt"
	"forum/database"
	"forum/handlers"

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

	fmt.Println("server started on http://localhost:1216")
	http.ListenAndServe(":1216", nil)

	// You can now use the 'db' connection to perform database operations.

}
