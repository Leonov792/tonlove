const { Pool } = require("pg");
const crypto = require("crypto");

const pool = new Pool({
  connectionString: process.env.DATABASE_URL || "postgresql://neondb_owner:npg_q3zRB4hiGEsk@ep-calm-bar-atzg09ti.c-9.us-east-1.aws.neon.tech/neondb?sslmode=require",
  ssl: { rejectUnauthorized: false },
});

const giftShop = [
  { type: "rose_red", name: "Красная роза", emoji: "🌹", price: 1 },
  { type: "rose_pink", name: "Розовая роза", emoji: "🌸", price: 2 },
  { type: "bouquet", name: "Букет роз", emoji: "💐", price: 5 },
  { type: "heart", name: "Сердце", emoji: "❤️", price: 3 },
  { type: "diamond", name: "Бриллиант", emoji: "���", price: 10 },
  { type: "crown", name: "Корона", emoji: "👑", price: 15 },
  { type: "ring", name: "Кольцо", emoji: "💍", price: 20 },
  { type: "stars", name: "Звёзды", emoji: "✨", price: 7 },
];

async function migrate() {
  await pool.query(`
    CREATE TABLE IF NOT EXISTS users (
      id BIGINT PRIMARY KEY, username TEXT, first_name TEXT DEFAULT '', age INTEGER DEFAULT 18,
      city TEXT DEFAULT '', bio TEXT DEFAULT '', photo_url TEXT DEFAULT '', balance INTEGER DEFAULT 10,
      created_at TIMESTAMP DEFAULT NOW()
    )`);
  await pool.query(`CREATE TABLE IF NOT EXISTS likes (id SERIAL PRIMARY KEY, from_user BIGINT REFERENCES users(id), to_user BIGINT REFERENCES users(id), created_at TIMESTAMP DEFAULT NOW(), UNIQUE(from_user, to_user))`);
  await pool.query(`CREATE TABLE IF NOT EXISTS gifts (id SERIAL PRIMARY KEY, from_user BIGINT REFERENCES users(id), to_user BIGINT REFERENCES users(id), gift_type TEXT, gift_name TEXT, price INTEGER, message TEXT DEFAULT '', created_at TIMESTAMP DEFAULT NOW())`);
  await pool.query(`INSERT INTO users (id, username, first_name, age, city, bio) VALUES (111,'anna_love','Анна',24,'Москва','Люблю путешествия и кофе ☕'),(222,'dmitry_m','Дмитрий',27,'СПб','Спорт, музыка, IT'),(333,'elena_s','Елена',22,'Казань','Творческая душа 🎨'),(444,'alex_k','Алексей',29,'Москва','Бизнес, авто, путешествия'),(555,'maria_v','Мария',25,'Новосибирск','Йога, книги, природа 🌿') ON CONFLICT (id) DO NOTHING`);
}

migrate().catch(console.error);

function json(res, data, status = 200) {
  res.setHeader("Content-Type", "application/json");
  res.setHeader("Access-Control-Allow-Origin", "*");
  res.setHeader("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS");
  res.setHeader("Access-Control-Allow-Headers", "*");
  res.status(status).send(JSON.stringify(data));
}

function getUserId(req) {
  // Parse from headers, then from query string  
  const fromHeader = parseInt(req.headers["x-telegram-user-id"] || "0");
  if (fromHeader > 0) return fromHeader;
  
  // Parse query string manually (Vercel doesn't set req.query for rewrites)
  const qs = (req.url || "").split("?")[1] || "";
  const params = new URLSearchParams(qs);
  return parseInt(params.get("user_id") || "0");
}

module.exports = async (req, res) => {
  if (req.method === "OPTIONS") return json(res, { ok: true });

  // Vercel rewrites strip the prefix; use original URL from header
  let path = req.headers["x-vercel-original-url"] || req.url;
  if (!path) path = req.url;
  path = path.replace(/\?.*/, "").replace(/\/+$/, "").replace(/^\/api/, "");
  
  const parts = path.split("/").filter(Boolean);
  const uid = getUserId(req);
  
  console.log("Request:", req.method, path, "UID:", uid);

  try {
    if (path.includes("/health")) return json(res, { status: "ok", db: !!pool });

    if (path.includes("/auth")) {
      const u = req.body;
      await pool.query("INSERT INTO users (id,username,first_name) VALUES ($1,$2,$3) ON CONFLICT (id) DO UPDATE SET username=$2,first_name=$3", [u.id, u.username, u.first_name]);
      return json(res, { success: true });
    }

    if (!uid) return json(res, { error: "Unauthorized" }, 401);

    if (path === "/profiles" || path === "/profiles/") {
      const { rows } = await pool.query("SELECT id,username,first_name,age,city,bio,photo_url,balance FROM users WHERE id!=$1 AND id NOT IN (SELECT to_user FROM likes WHERE from_user=$1) ORDER BY RANDOM() LIMIT 20", [uid]);
      return json(res, rows);
    }

    if (path.includes("/like")) {
      const tid = parseInt(parts[parts.length - 2] || parts[parts.length - 1]);
      await pool.query("INSERT INTO likes (from_user,to_user) VALUES ($1,$2) ON CONFLICT DO NOTHING", [uid, tid]);
      const { rows } = await pool.query("SELECT EXISTS(SELECT 1 FROM likes WHERE from_user=$1 AND to_user=$2)", [tid, uid]);
      return json(res, { liked: true, match: rows[0].exists });
    }

    if (path.includes("/skip")) return json(res, { skipped: true });

    if (path === "/profile" && req.method === "GET") {
      const { rows } = await pool.query("SELECT * FROM users WHERE id=$1", [uid]);
      return json(res, rows[0] || { error: "not found" });
    }

    if (path === "/profile" && req.method === "PUT") {
      const u = req.body;
      await pool.query("UPDATE users SET first_name=$1,age=$2,city=$3,bio=$4 WHERE id=$5", [u.first_name, u.age, u.city, u.bio, uid]);
      return json(res, { success: true });
    }

    if (path.includes("/gifts/shop")) return json(res, giftShop);

    if (path.includes("/gifts/send")) {
      const g = req.body;
      const { rows } = await pool.query("SELECT balance FROM users WHERE id=$1", [uid]);
      if (rows[0].balance < g.price) return json(res, { error: "Недостаточно роз" }, 400);
      await pool.query("UPDATE users SET balance=balance-$1 WHERE id=$2", [g.price, uid]);
      await pool.query("INSERT INTO gifts (from_user,to_user,gift_type,gift_name,price,message) VALUES ($1,$2,$3,$4,$5,$6)", [uid, g.to_user, g.gift_type, g.gift_name, g.price, g.message]);
      return json(res, { success: true });
    }

    if (path.includes("/gifts/received")) {
      const { rows } = await pool.query("SELECT g.*,u.first_name as from_name FROM gifts g JOIN users u ON g.from_user=u.id WHERE g.to_user=$1 ORDER BY g.created_at DESC LIMIT 50", [uid]);
      return json(res, rows);
    }

    if (path.includes("/gifts/sent")) {
      const { rows } = await pool.query("SELECT g.*,u.first_name as to_name FROM gifts g JOIN users u ON g.to_user=u.id WHERE g.from_user=$1 ORDER BY g.created_at DESC LIMIT 50", [uid]);
      return json(res, rows);
    }

    if (path.includes("/matches")) {
      const { rows } = await pool.query("SELECT u.* FROM users u WHERE u.id!=$1 AND u.id IN (SELECT to_user FROM likes WHERE from_user=$1) AND u.id IN (SELECT from_user FROM likes WHERE to_user=$1)", [uid]);
      return json(res, rows);
    }

    if (path.includes("/notifications")) return json(res, []);

    json(res, { status: "ok", message: "TON Love API v1" });
  } catch (e) {
    console.error(e);
    json(res, { error: "Internal error" }, 500);
  }
};
