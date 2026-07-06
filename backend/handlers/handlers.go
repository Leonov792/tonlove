package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"tonlove-backend/middleware"
	"tonlove-backend/models"

	"github.com/gorilla/mux"
)

type Handler struct{ db *sql.DB }

func New(db *sql.DB) *Handler { return &Handler{db} }

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// POST /api/auth — register/login via Telegram
func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	_, err := h.db.Exec(`
		INSERT INTO users (id, username, first_name) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET username=$2, first_name=$3`,
		body.ID, body.Username, body.FirstName)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "user_id": body.ID})
}

// GET /api/profiles — get swipeable profiles
func (h *Handler) GetProfiles(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	rows, err := h.db.Query(`
		SELECT id, username, first_name, age, city, bio, photo_url, balance, created_at
		FROM users
		WHERE id != $1
		AND id NOT IN (SELECT to_user FROM likes WHERE from_user = $1)
		ORDER BY RANDOM()
		LIMIT 20`, userID)
	if err != nil {
		http.Error(w, `[]`, http.StatusOK)
		return
	}
	defer rows.Close()

	var profiles []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.FirstName, &u.Age, &u.City, &u.Bio, &u.PhotoURL, &u.Balance, &u.CreatedAt)
		profiles = append(profiles, u)
	}
	if profiles == nil {
		profiles = []models.User{}
	}
	json.NewEncoder(w).Encode(profiles)
}

// POST /api/profiles/{id}/like
func (h *Handler) LikeProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	targetID, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)

	h.db.Exec(`INSERT INTO likes (from_user, to_user) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, targetID)

	// Check mutual like
	var isMatch bool
	h.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM likes WHERE from_user=$1 AND to_user=$2)`, targetID, userID).Scan(&isMatch)

	if isMatch {
		h.db.Exec(`INSERT INTO notifications (user_id, msg_type, message, from_user) VALUES ($1,'match','У вас взаимная симпатия! 💕',$2)`, userID, targetID)
		h.db.Exec(`INSERT INTO notifications (user_id, msg_type, message, from_user) VALUES ($1,'match','У вас взаимная симпатия! 💕',$2)`, targetID, userID)
		json.NewEncoder(w).Encode(map[string]interface{}{"liked": true, "match": true})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{"liked": true, "match": false})
	}
}

// POST /api/profiles/{id}/skip
func (h *Handler) SkipProfile(w http.ResponseWriter, r *http.Request) {
	targetID := mux.Vars(r)["id"]
	json.NewEncoder(w).Encode(map[string]interface{}{"skipped": true, "id": targetID})
}

// GET /api/profile
func (h *Handler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var u models.User
	err := h.db.QueryRow(`
		SELECT id, username, first_name, age, city, bio, photo_url, balance, created_at
		FROM users WHERE id=$1`, userID).
		Scan(&u.ID, &u.Username, &u.FirstName, &u.Age, &u.City, &u.Bio, &u.PhotoURL, &u.Balance, &u.CreatedAt)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(u)
}

// PUT /api/profile
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body struct {
		FirstName string `json:"first_name"`
		Age       int    `json:"age"`
		City      string `json:"city"`
		Bio       string `json:"bio"`
		PhotoURL  string `json:"photo_url"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	h.db.Exec(`UPDATE users SET first_name=$1, age=$2, city=$3, bio=$4, photo_url=$5 WHERE id=$6`,
		body.FirstName, body.Age, body.City, body.Bio, body.PhotoURL, userID)

	w.Write([]byte(`{"success":true}`))
}

// GET /api/gifts/shop
func (h *Handler) GetGiftShop(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(models.GiftShop)
}

// POST /api/gifts/send
func (h *Handler) SendGift(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	var body struct {
		ToUser  int64  `json:"to_user"`
		GiftType string `json:"gift_type"`
		GiftName string `json:"gift_name"`
		Price    int    `json:"price"`
		Message  string `json:"message"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	// Check balance
	var balance int
	h.db.QueryRow(`SELECT balance FROM users WHERE id=$1`, userID).Scan(&balance)
	if balance < body.Price {
		http.Error(w, `{"error":"Недостаточно роз 💎"}`, http.StatusBadRequest)
		return
	}

	tx, _ := h.db.Begin()
	tx.Exec(`UPDATE users SET balance = balance - $1 WHERE id = $2`, body.Price, userID)
	tx.Exec(`INSERT INTO gifts (from_user, to_user, gift_type, gift_name, price, message) VALUES ($1,$2,$3,$4,$5,$6)`,
		userID, body.ToUser, body.GiftType, body.GiftName, body.Price, body.Message)
	tx.Exec(`INSERT INTO notifications (user_id, msg_type, message, from_user) VALUES ($1,'gift',$2,$3)`,
		body.ToUser, "Вам отправили подарок: "+body.GiftName+" 🎁", userID)
	tx.Commit()

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "balance_left": balance - body.Price})
}

// GET /api/gifts/received
func (h *Handler) GetReceivedGifts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	rows, _ := h.db.Query(`
		SELECT g.id, g.from_user, g.to_user, g.gift_type, g.gift_name, g.price, g.message, g.created_at,
			u.first_name, u.username
		FROM gifts g JOIN users u ON g.from_user = u.id
		WHERE g.to_user = $1 ORDER BY g.created_at DESC LIMIT 50`, userID)
	defer rows.Close()

	var gifts []map[string]interface{}
	for rows.Next() {
		var g models.Gift
		var fromName, fromUsername string
		var date time.Time
		rows.Scan(&g.ID, &g.FromUser, &g.ToUser, &g.GiftType, &g.GiftName, &g.Price, &g.Message, &date, &fromName, &fromUsername)
		gifts = append(gifts, map[string]interface{}{
			"id": g.ID, "from_user": g.FromUser, "gift_type": g.GiftType,
			"gift_name": g.GiftName, "price": g.Price, "message": g.Message,
			"from_name": fromName, "date": date,
		})
	}
	if gifts == nil { gifts = []map[string]interface{}{} }
	json.NewEncoder(w).Encode(gifts)
}

// GET /api/gifts/sent
func (h *Handler) GetSentGifts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	rows, _ := h.db.Query(`
		SELECT g.id, g.to_user, g.gift_type, g.gift_name, g.price, g.message, g.created_at, u.first_name
		FROM gifts g JOIN users u ON g.to_user = u.id
		WHERE g.from_user = $1 ORDER BY g.created_at DESC LIMIT 50`, userID)
	defer rows.Close()

	var gifts []map[string]interface{}
	for rows.Next() {
		var g models.Gift
		var toName string
		var date time.Time
		rows.Scan(&g.ID, &g.ToUser, &g.GiftType, &g.GiftName, &g.Price, &g.Message, &date, &toName)
		gifts = append(gifts, map[string]interface{}{
			"id": g.ID, "to_user": g.ToUser, "gift_type": g.GiftType,
			"gift_name": g.GiftName, "price": g.Price, "message": g.Message,
			"to_name": toName, "date": date,
		})
	}
	if gifts == nil { gifts = []map[string]interface{}{} }
	json.NewEncoder(w).Encode(gifts)
}

// GET /api/matches
func (h *Handler) GetMatches(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	rows, _ := h.db.Query(`
		SELECT u.id, u.username, u.first_name, u.age, u.city, u.bio, u.photo_url
		FROM users u
		WHERE u.id != $1
		AND u.id IN (SELECT to_user FROM likes WHERE from_user = $1)
		AND u.id IN (SELECT from_user FROM likes WHERE to_user = $1)`, userID)
	defer rows.Close()

	var matches []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.FirstName, &u.Age, &u.City, &u.Bio, &u.PhotoURL)
		matches = append(matches, u)
	}
	if matches == nil { matches = []models.User{} }
	json.NewEncoder(w).Encode(matches)
}

// GET /api/notifications
func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	rows, _ := h.db.Query(`
		SELECT id, msg_type, message, from_user, is_read, created_at
		FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 30`, userID)
	defer rows.Close()

	var notifs []models.Notification
	for rows.Next() {
		var n models.Notification
		rows.Scan(&n.ID, &n.MsgType, &n.Message, &n.FromUser, &n.IsRead, &n.Date)
		notifs = append(notifs, n)
	}
	if notifs == nil { notifs = []models.Notification{} }
	json.NewEncoder(w).Encode(notifs)
}
