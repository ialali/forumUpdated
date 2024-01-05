package database

import (
	"database/sql"
	"fmt"
	"log"

	auth "forum/middleware"
	"time"
)

func RegisterUser(db *sql.DB, username, email, password string) (int64, error) {
	// Hash the password before inserting it into the database (assuming you've set up bcrypt).
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return 0, err
	}

	// Get the current registration date.
	registrationDate := time.Now().Format("2006-01-02 15:04:05")

	result, err := db.Exec(`
        INSERT INTO users (username, email, password, registration_date)
        VALUES (?, ?, ?, ?);
    `, &username, &email, &hashedPassword, registrationDate)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func UserExists(db *sql.DB, username string) bool {

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		log.Println("Error checking if user exists:", err)
		return false
	}

	return count > 0
}

func GetUserByID(db *sql.DB, userID int) (User, error) {
	var user User
	err := db.QueryRow(`
		SELECT id, username, email, password, registration_date
		FROM users
		WHERE id = ?;
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.RegistrationDate)

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func GetUserByEmail(db *sql.DB, email string) (User, error) {
	var user User
	err := db.QueryRow(`SELECT id, email, password, registration_date FROM users WHERE email = ?`, email).Scan(&user.ID, &user.Email, &user.Username, &user.RegistrationDate)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
func GetIDBYusername(db *sql.DB, username string) (int, error) {
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// CreatePost inserts a new post into the database and returns the post ID.
func InsertPost(db *sql.DB, category, title, content string, userID int) error {
	// Prepare the SQL statement to insert a new post.
	stmt, err := db.Prepare("INSERT INTO posts (user_id, title, content, category, created_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Get the current timestamp.
	createdAt := time.Now().Format("2006-01-02 15:04:05")

	// Execute the SQL statement to insert the new post.
	_, err = stmt.Exec(userID, title, content, category, createdAt)
	if err != nil {
		return err
	}

	return nil
}

func GetPosts(db *sql.DB) ([]Post, error) {
	var posts []Post

	rows, err := db.Query("SELECT posts.id, posts.title, posts.content, posts.category, posts.created_at, users.username FROM posts INNER JOIN users ON posts.user_id = users.id ORDER BY posts.id DESC;")

	if err != nil {
		fmt.Println("error querrying database")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post

		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category, &post.CreationDate, &post.Username); err != nil {

			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func InsertComment(db *sql.DB, postID, userID int, content string) error {
	_, err := db.Exec("INSERT INTO comments (user_id, post_id, content, creation_date) VALUES (?, ?, ?, ?)",
		userID, postID, content, time.Now().Format(time.DateTime))
	return err
}
func GetCommentsForPost(db *sql.DB, postID int) ([]Comment, error) {
	var comments []Comment
	rows, err := db.Query("SELECT comments.id, comments.post_id, comments.user_id, comments.content, comments.creation_date, users.username FROM comments INNER JOIN users ON comments.user_id = users.id WHERE comments.post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreationDate, &comment.Username)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

func InsertPostLike(db *sql.DB, userID, postID int) error {
	var likeStatus bool

	err := db.QueryRow("SELECT like_status FROM post_likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&likeStatus)
	if err == sql.ErrNoRows {
		_, err := db.Exec("INSERT INTO post_likes (user_id, post_id, like_status, like) VALUES (?, ?, ?, ?)", userID, postID, true, 1)
		return err
	} else if err != nil {
		return err
	} else {
		if likeStatus {
			_, err := db.Exec("DELETE FROM post_likes WHERE user_id = ? AND post_id = ?", userID, postID)
			return err
		} else {
			_, err := db.Exec("UPDATE post_likes SET like_status = ?, like = like + 1 WHERE user_id = ? AND post_id = ?", true, userID, postID)
			return err
		}
	}
}
func InsertPostDislike(db *sql.DB, userID, postID int) error {
	var likeStatus bool

	err := db.QueryRow("SELECT like_status FROM post_likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&likeStatus)
	if err == sql.ErrNoRows {
		_, err := db.Exec("INSERT INTO post_likes (user_id, post_id, like_status, like) VALUES (?, ?, ?, ?)", userID, postID, false, -1)
		return err
	} else if err != nil {
		return err
	} else {
		if likeStatus {
			_, err := db.Exec("DELETE FROM post_likes WHERE user_id = ? AND post_id = ?", userID, postID)
			return err
		} else {
			_, err := db.Exec("UPDATE post_likes SET like_status = ?, like = like + 1 WHERE user_id = ? AND post_id = ?", false, userID, postID)
			return err
		}
	}
}

func GetPostLikesCount(db *sql.DB, postID int) (int, int, error) {
	var likeCount, dislikeCount int

	// Query for likes count
	err := db.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND like_status = ?", postID, true).Scan(&likeCount)
	if err != nil {
		return 0, 0, err
	}

	// Query for dislikes count
	err = db.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND like_status = ?", postID, false).Scan(&dislikeCount)
	if err != nil {
		return 0, 0, err
	}

	return likeCount, dislikeCount, nil
}

func InsertCommentLike(db *sql.DB, userID, commentID int) error {
	var likeStatus bool

	err := db.QueryRow("SELECT like_status FROM comments_likes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&likeStatus)
	if err == sql.ErrNoRows {
		_, err := db.Exec("INSERT INTO comments_likes (user_id, comment_id, like_status, like) VALUES (?, ?, ?, ?)", userID, commentID, true, 1)
		return err
	} else if err != nil {
		return err
	} else {
		if likeStatus {
			_, err := db.Exec("DELETE FROM comments_likes WHERE user_id = ? AND comment_id = ?", userID, commentID)
			return err
		} else {
			_, err := db.Exec("UPDATE comments_likes SET like_status = ?, like = like + 1 WHERE user_id = ? AND comment_id = ?", true, userID, commentID)
			return err
		}
	}
}

func InsertCommentDislike(db *sql.DB, userID, commentID int) error {
	var likeStatus bool

	err := db.QueryRow("SELECT like_status FROM comments_likes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&likeStatus)
	if err == sql.ErrNoRows {
		_, err := db.Exec("INSERT INTO comments_likes (user_id, comment_id, like_status, like) VALUES (?, ?, ?, ?)", userID, commentID, false, -1)
		return err
	} else if err != nil {
		return err
	} else {
		if likeStatus {
			_, err := db.Exec("DELETE FROM comments_likes WHERE user_id = ? AND comment_id = ?", userID, commentID)
			return err
		} else {
			_, err := db.Exec("UPDATE comments_likes SET like_status = ?, like = like + 1 WHERE user_id = ? AND comment_id = ?", false, userID, commentID)
			return err
		}
	}
}

func GetCommentLikesCount(db *sql.DB, postID int) (int, int, error) {
	var likeCount, dislikeCount int

	// Query for likes count
	err := db.QueryRow("SELECT COUNT(*) FROM comments_likes WHERE comment_id = ? AND like_status = ?", postID, true).Scan(&likeCount)
	if err != nil {
		return 0, 0, err
	}

	// Query for dislikes count
	err = db.QueryRow("SELECT COUNT(*) FROM comments_likes WHERE comment_id = ? AND like_status = ?", postID, false).Scan(&dislikeCount)
	if err != nil {
		return 0, 0, err
	}

	return likeCount, dislikeCount, nil
}
func GetOwnedPosts(db *sql.DB, userID int) ([]Post, error) {
	query := `
        SELECT 
            posts.id, posts.user_id, posts.title, posts.content, posts.category, users.username 
        FROM 
            posts 
        INNER JOIN 
            users 
        ON 
            posts.user_id = users.id 
        WHERE 
            posts.user_id = ?
    `
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Category, &post.Username); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func GetLikedPosts(db *sql.DB, userID int, liked bool) ([]Post, error) {
	query := `
        SELECT 
            p.id, p.user_id, p.title, p.content, p.category, u.username
        FROM 
            posts p
        INNER JOIN 
            post_likes pl ON p.id = pl.post_id
        INNER JOIN 
            users u ON p.user_id = u.id
        WHERE 
            pl.user_id = ? AND pl.like_status = ?
    `

	rows, err := db.Query(query, userID, liked)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var likedPosts []Post

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Category, &post.Username); err != nil {
			return nil, err
		}
		likedPosts = append(likedPosts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return likedPosts, nil
}

func GetPostsByCategory(db *sql.DB, category string, userID int, createdByMe, likedByMe bool) ([]Post, error) {
	var query string
	var args []interface{}

	// Base query
	query = "SELECT posts.id, posts.user_id, posts.title, posts.content, posts.category, users.username FROM posts INNER JOIN users ON posts.user_id = users.id WHERE posts.category = ?"
	args = append(args, category)

	if createdByMe {
		// Add condition to filter by user ID
		query += " AND posts.user_id = ?"
		args = append(args, userID)
	}

	if likedByMe {
		// Add condition to filter by liked posts
		query += " AND posts.id IN (SELECT post_id FROM post_likes WHERE user_id = ?)"
		args = append(args, userID)
	}

	// Use Query instead of QueryRow
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Category, &post.Username)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Error iterating over rows:", err)
		return nil, err
	}

	return posts, nil
}
