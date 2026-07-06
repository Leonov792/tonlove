package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func Connect(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	log.Println("PostgreSQL connected")
	return db, nil
}

func Migrate(db *sql.DB) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT PRIMARY KEY,
			username TEXT,
			first_name TEXT DEFAULT '',
			age INTEGER DEFAULT 18,
			city TEXT DEFAULT '',
			bio TEXT DEFAULT '',
			photo_url TEXT DEFAULT '',
			balance INTEGER DEFAULT 10,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS likes (
			id SERIAL PRIMARY KEY,
			from_user BIGINT REFERENCES users(id),
			to_user BIGINT REFERENCES users(id),
			created_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(from_user, to_user)
		)`,
		`CREATE TABLE IF NOT EXISTS gifts (
			id SERIAL PRIMARY KEY,
			from_user BIGINT REFERENCES users(id),
			to_user BIGINT REFERENCES users(id),
			gift_type TEXT NOT NULL,
			gift_name TEXT NOT NULL,
			price INTEGER NOT NULL,
			message TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS notifications (
			id SERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES users(id),
			msg_type TEXT NOT NULL,
			message TEXT NOT NULL,
			from_user BIGINT REFERENCES users(id),
			is_read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`INSERT INTO users (id, username, first_name, age, city, bio, photo_url) VALUES 
			(111, 'anna_love', 'Анна', 24, 'Москва', 'Люблю путешествия и кофе ☕', ''),
			(222, 'dmitry_m', 'Дмитрий', 27, 'СПб', 'Спорт, музыка, IT', ''),
			(333, 'elena_s', 'Елена', 22, 'Казань', 'Творческая душа, ищу музу 🎨', ''),
			(444, 'alex_k', 'Алексей', 29, 'Москва', 'Бизнес, авто, путешествия', ''),
			(555, 'maria_v', 'Мария', 25, 'Новосибирск', 'Йога, книги, природа 🌿', '')
		ON CONFLICT (id) DO NOTHING`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Printf("Migration warning: %v", err)
		}
	}
	log.Println("Database migrated")
}
