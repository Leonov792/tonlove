package main

import (
	"log"
	"net/http"
	"os"

	"tonlove-backend/database"
	"tonlove-backend/handlers"
	"tonlove-backend/middleware"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/tonlove?sslmode=disable"
	}

	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	database.Migrate(db)

	h := handlers.New(db)

	r := mux.NewRouter()

	// Public
	r.HandleFunc("/api/health", h.Health).Methods("GET")
	r.HandleFunc("/api/auth", h.Auth).Methods("POST")

	// Protected
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.Auth)

	api.HandleFunc("/profiles", h.GetProfiles).Methods("GET")
	api.HandleFunc("/profiles/{id}/like", h.LikeProfile).Methods("POST")
	api.HandleFunc("/profiles/{id}/skip", h.SkipProfile).Methods("POST")
	api.HandleFunc("/profile", h.GetMyProfile).Methods("GET")
	api.HandleFunc("/profile", h.UpdateProfile).Methods("PUT")

	api.HandleFunc("/gifts/shop", h.GetGiftShop).Methods("GET")
	api.HandleFunc("/gifts/send", h.SendGift).Methods("POST")
	api.HandleFunc("/gifts/received", h.GetReceivedGifts).Methods("GET")
	api.HandleFunc("/gifts/sent", h.GetSentGifts).Methods("GET")

	api.HandleFunc("/matches", h.GetMatches).Methods("GET")
	api.HandleFunc("/notifications", h.GetNotifications).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("TON Love API running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, c.Handler(r)))
}
