package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() {
	var err error
	DB, err = sql.Open("sqlite3", "./meowmico.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	createTables()
	log.Println("Database connected")
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS events (
		id			INTEGER PRIMARY KEY AUTOINCREMENT,
		channel		TEXT NOT NULL,
		content		TEXT NOT NULL,
		created_at 	DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
		id			INTEGER PRIMARY KEY AUTOINCREMENT,
		role		TEXT NOT NULL,
		content		TEXT NOT NULL,
		created_at 	DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS preferences (
		key			TEXT PRIMARY KEY,
		value		TEXT NOT NULL
		)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			log.Fatal("Failed to create table:", err)
		}
	}

	log.Println("Tables ready <3 ")
}
