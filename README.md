# 💎 TON Love — Telegram Mini App для дейтинга

Telegram Mini App для знакомств через кружки TON. Карточки анкет, лайки, взаимные симпатии, магазин подарков.

## 🛠 Tech Stack

| Слой | Технология |
|------|-----------|
| **Frontend** | HTML, CSS, JavaScript (Telegram WebApp SDK) |
| **Backend** | Go 1.21, Gorilla Mux |
| **Database** | PostgreSQL 16 |
| **Auth** | Telegram WebApp initData |
| **Deploy** | Docker + docker-compose |

## 🚀 Quick Start

### Local (Docker)
```bash
docker-compose up -d
# API: http://localhost:8080
# Frontend: http://localhost:3000
```

### Manual
```bash
# Database
createdb tonlove

# Backend
cd backend
go mod download
DATABASE_URL="postgres://postgres:password@localhost:5432/tonlove?sslmode=disable" go run .

# Frontend
cd frontend
python -m http.server 3000
```

## 📡 API Endpoints

### Auth
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth` | Register/login via Telegram |
| GET | `/api/health` | Health check |

### Profiles
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/profiles` | Get swipeable profiles |
| POST | `/api/profiles/{id}/like` | Like a profile |
| POST | `/api/profiles/{id}/skip` | Skip a profile |
| GET | `/api/profile` | Get my profile |
| PUT | `/api/profile` | Update my profile |

### Gifts
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/gifts/shop` | Gift shop items |
| POST | `/api/gifts/send` | Send a gift |
| GET | `/api/gifts/received` | Received gifts |
| GET | `/api/gifts/sent` | Sent gifts |

### Social
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/matches` | Mutual likes |
| GET | `/api/notifications` | Notifications |

## 📱 Telegram Bot Setup

1. Create bot via [@BotFather](https://t.me/BotFather)
2. Set WebApp URL to your frontend deployment
3. In `js/app.js`, update `API` variable to your backend URL

## 🎨 Features

- 💕 Swipe cards (Tinder-like) with animations
- 💎 Gift shop (virtual roses, hearts, diamonds)
- 💞 Mutual matches detection
- 🔔 Real-time notifications (polling)
- 👤 Profile editing
- 📱 Adaptive mobile-first design

## 📸 Screenshots

*(Open in Telegram via @YourBot)*

## 📄 License

MIT
