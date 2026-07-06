package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const token = "8650667280:AAE7NbQYdXbZg8SmY3WTv_rPXL4WCJ9YEWY"
const baseURL = "https://api.telegram.org/bot" + token

// Set WebApp menu button for the bot
func main() {
	webAppURL := os.Getenv("WEBAPP_URL")
	if webAppURL == "" {
		webAppURL = "https://tonlove.vercel.app"
	}

	// Set menu button
	if err := setMenuButton(webAppURL); err != nil {
		log.Printf("Menu button warning: %v", err)
	} else {
		log.Println("Menu button set to:", webAppURL)
	}

	// Start polling for /start command
	log.Println("Bot @SergGOrelyyBot starting...")
	offset := 0
	for {
		updates, err := getUpdates(offset)
		if err != nil {
			log.Printf("GetUpdates error: %v", err)
			continue
		}
		for _, u := range updates {
			offset = u.UpdateID + 1
			if u.Message != nil && u.Message.Text == "/start" {
				sendMessage(u.Message.Chat.ID,
					"💎 Добро пожаловать в **TON Love**!\n\n"+
						"Знакомься с новыми людьми, дари подарки и находи свою любовь!\n\n"+
						"Нажми кнопку ниже чтобы открыть приложение 👇",
					webAppURL)
			}
		}
	}
}

type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
}
type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text,omitempty"`
}
type Chat struct {
	ID int64 `json:"id"`
}

func getUpdates(offset int) ([]Update, error) {
	url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=30", baseURL, offset)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
	json.Unmarshal(body, &result)
	return result.Result, nil
}

func sendMessage(chatID int64, text, webAppURL string) error {
	keyboard := map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "💕 Открыть TON Love", "web_app": map[string]string{"url": webAppURL}}},
		},
	}
	body := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "Markdown",
		"reply_markup": keyboard,
	}
	data, _ := json.Marshal(body)
	resp, err := http.Post(baseURL+"/sendMessage", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func setMenuButton(webAppURL string) error {
	body := map[string]interface{}{
		"menu_button": map[string]interface{}{
			"type":    "web_app",
			"text":    "💕 TON Love",
			"web_app": map[string]string{"url": webAppURL},
		},
	}
	data, _ := json.Marshal(body)
	resp, err := http.Post(baseURL+"/setChatMenuButton", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	log.Println("setMenuButton response:", string(respBody))
	return nil
}
