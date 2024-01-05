package handlers

import (
	"database/sql"
	"forum/database"
	"log"
	"net/http"
	"text/template"
)

func ShowPostHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userData := GetAuthenticatedUserData(db, r)
	posts, err := database.GetPosts(db)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	for i, post := range posts {
		comments, err := database.GetCommentsForPost(db, post.ID)
		if err != nil {
			http.Error(w, "Error fetching comments", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		likeCount, dislikeCount, err := database.GetPostLikesCount(db, post.ID)
		if err != nil {
			http.Error(w, "Error fetching post likes/dislikes", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		for j, comment := range comments {
			commentLikeCount, commentDislikeCount, err := database.GetCommentLikesCount(db, comment.ID)
			if err != nil {
				http.Error(w, "Error fetching comment likes/dislikes", http.StatusInternalServerError)
				log.Println(err)
				return
			}

			comments[j].LikeCount = commentLikeCount
			comments[j].DislikeCount = commentDislikeCount
		}

		post.Comments = comments
		post.LikeCount = likeCount
		post.DislikeCount = dislikeCount
		posts[i] = post
	}

	data := database.PageData{
		IsAuthenticated: userData.IsAuthenticated,
		Username:        userData.Username,
		Posts:           posts,
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Error Parsing index.html", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}
