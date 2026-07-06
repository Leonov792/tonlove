package models

import "time"

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	Age       int       `json:"age"`
	City      string    `json:"city"`
	Bio       string    `json:"bio"`
	PhotoURL  string    `json:"photo_url"`
	Balance   int       `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type Like struct {
	ID       int64     `json:"id"`
	FromUser int64     `json:"from_user"`
	ToUser   int64     `json:"to_user"`
	Date     time.Time `json:"date"`
}

type Gift struct {
	ID       int64     `json:"id"`
	FromUser int64     `json:"from_user"`
	ToUser   int64     `json:"to_user"`
	GiftType string    `json:"gift_type"`
	GiftName string    `json:"gift_name"`
	Price    int       `json:"price"`
	Message  string    `json:"message"`
	Date     time.Time `json:"date"`
}

type Notification struct {
	ID       int64     `json:"id"`
	UserID   int64     `json:"user_id"`
	MsgType  string    `json:"msg_type"`
	Message  string    `json:"message"`
	FromUser int64     `json:"from_user"`
	IsRead   bool      `json:"is_read"`
	Date     time.Time `json:"date"`
}

type ProfileResponse struct {
	User     User   `json:"user"`
	Distance string `json:"distance"`
}

type GiftShopItem struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Emoji string `json:"emoji"`
	Price int   `json:"price"`
}

var GiftShop = []GiftShopItem{
	{Type: "rose_red", Name: "Красная роза", Emoji: "🌹", Price: 1},
	{Type: "rose_pink", Name: "Розовая роза", Emoji: "🌸", Price: 2},
	{Type: "bouquet", Name: "Букет роз", Emoji: "💐", Price: 5},
	{Type: "heart", Name: "Сердце", Emoji: "❤️", Price: 3},
	{Type: "diamond", Name: "Бриллиант", Emoji: "💎", Price: 10},
	{Type: "crown", Name: "Корона", Emoji: "👑", Price: 15},
	{Type: "ring", Name: "Кольцо", Emoji: "💍", Price: 20},
	{Type: "stars", Name: "Звёзды", Emoji: "✨", Price: 7},
}
