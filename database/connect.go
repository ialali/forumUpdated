package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDatabase(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable foreign key constraints (optional, if you have foreign keys)
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, err
	}

	return db, nil
}

// InitializeSchema creates the database tables.
func InitializeSchema(db *sql.DB) error {
	// Create the 'users' table
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            username TEXT  NOT NULL,
            email TEXT UNIQUE NOT NULL UNIQUE,
            password TEXT NOT NULL,
            registration_date DATETIME
        );
    `)
	if err != nil {
		return err
	}

	// Create the 'posts' table with foreign key
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS posts (
            post_id INTEGER PRIMARY KEY,
            user_id INTEGER,
			title TEXT,
            content TEXT,
			category TEXT,
			created_at DATETIME,
            FOREIGN KEY (user_id) REFERENCES users(id)
        );
    `)
	if err != nil {
		return err
	}

	// Create the 'comments' table with foreign keys
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS comments (
            id INTEGER PRIMARY KEY,
            user_id INTEGER,
            post_id INTEGER,
            content TEXT,
            creation_date DATETIME,
            FOREIGN KEY (user_id) REFERENCES users(id),
            FOREIGN KEY (post_id) REFERENCES posts(id)
        );
    `)
	if err != nil {
		return err
	}

	// Create the 'categories' table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY,
            name TEXT
        );
    `)
	if err != nil {
		return err
	}

	// Create the 'likes' table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS post_likes (
            id INTEGER PRIMARY KEY,
            user_id INTEGER,
            post_id INTEGER,
            like_status BOOLEAN,
            like INTEGER DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (post_id) REFERENCES posts(id) 
        );
    `)
	if err != nil {
		return err
	}
	
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS comments_likes (
            id INTEGER PRIMARY KEY,
            user_id INTEGER,
            comment_id INTEGER,
            like_status BOOLEAN,
            like INTEGER DEFAULT 0,
            FOREIGN KEY (user_id) REFERENCES users(id),
            FOREIGN KEY (comment_id) REFERENCES comments(id)
        );
    `)
	if err != nil {
		return err
	}
	

	return nil
}
