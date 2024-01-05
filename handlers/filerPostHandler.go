package handlers

import (
	"database/sql"
	"forum/database"
	"log"
	"net/http"
	"text/template"
)

func FilterPosts(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// 1. Parse the filtering criteria from the request.
	category := r.FormValue("category")
	created := r.FormValue("created")
	liked := r.FormValue("liked")

	// Check if the user is authenticated
	userID, isAuthenticated := GetAuthenticatedUserID(r)
	if !isAuthenticated && (created == "true" || liked == "true") {
		// If not authenticated and trying to filter by ownership or liked posts, redirect to login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Initialize an empty slice to hold filtered posts.
	var posts []database.Post
	createdBool := created == "true"
	likedBool := liked == "true"

	// 2. Based on the criteria, call the corresponding functions to retrieve the filtered posts.
	switch {
	case category != "":
		// Filter by category for both authenticated and non-authenticated users
		posts, _ = database.GetPostsByCategory(db, category, userID, createdBool, likedBool)

		// If 'created' is true, filter created posts for authenticated users
		if created == "true" && isAuthenticated {
			userPosts, _ := database.GetOwnedPosts(db, userID)
			posts = intersection(posts, userPosts)
		}

		// If 'liked' is true, filter liked posts for authenticated users
		if liked == "true" && isAuthenticated {
			likedPosts, _ := database.GetLikedPosts(db, userID, true)
			posts = intersection(posts, likedPosts)
		}

	default:
		// No category selected, show all posts for both authenticated and non-authenticated users
		posts, _ = database.GetPosts(db)

		// If 'created' is true, filter created posts for authenticated users
		if created == "true" && isAuthenticated {
			userPosts, _ := database.GetOwnedPosts(db, userID)
			posts = intersection(posts, userPosts)
		}

		// If 'liked' is true, filter liked posts for authenticated users
		if liked == "true" && isAuthenticated {
			likedPosts, _ := database.GetLikedPosts(db, userID, true)
			posts = intersection(posts, likedPosts)
		}
	}

	// 3. For each post, retrieve like/dislike counts and usernames for comments.
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

		// Fetch like/dislike counts and usernames for comments
		for j := range comments {
			likeCount, dislikeCount, err := database.GetCommentLikesCount(db, comments[j].ID)
			if err != nil {
				http.Error(w, "Error fetching comment likes/dislikes", http.StatusInternalServerError)
				log.Println(err)
				return
			}

			comments[j].LikeCount = likeCount
			comments[j].DislikeCount = dislikeCount
		}

		// Assign comments to the post
		post.Comments = comments
		post.LikeCount = likeCount
		post.DislikeCount = dislikeCount
		posts[i] = post
	}

	// 4. Render the filtered posts to the page.
	userData := GetAuthenticatedUserData(db, r)

	data := struct {
		IsAuthenticated bool
		Username        string
		Posts           []database.Post
	}{
		IsAuthenticated: userData.IsAuthenticated,
		Username:        userData.Username,
		Posts:           posts,
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Error Parsing index.html", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering the template", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

// intersection returns the intersection of two slices of posts.
func intersection(a, b []database.Post) []database.Post {
	set := make(map[int]bool)
	var result []database.Post

	for _, post := range a {
		set[post.ID] = true
	}

	for _, post := range b {
		if set[post.ID] {
			result = append(result, post)
		}
	}

	return result
}
