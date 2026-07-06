// TON Love — Telegram Mini App
// ============================================

const API = 'https://backend-six-chi-80.vercel.app/api';
const tg = window.Telegram?.WebApp;

let currentUser = null;
let profiles = [];
let currentIndex = 0;
let currentGift = null;

// Init
document.addEventListener('DOMContentLoaded', () => {
    if (tg) {
        tg.ready();
        tg.expand();
    }
    initAuth();
    initTabs();
    initSwipe();
    initGifts();
    initProfile();
    initNotifications();
    loadProfiles();
});

async function api(path, method = 'GET', body = null) {
    const headers = { 'Content-Type': 'application/json' };
    if (currentUser) headers['X-Telegram-User-ID'] = currentUser.id;
    const opts = { method, headers };
    if (body) opts.body = JSON.stringify(body);
    const res = await fetch(API + path, opts);
    return res.ok ? res.json() : null;
}

function toast(msg) {
    const t = document.getElementById('toast');
    t.textContent = msg; t.classList.add('show');
    setTimeout(() => t.classList.remove('show'), 2000);
}

// ========== AUTH ==========
async function initAuth() {
    if (tg?.initDataUnsafe?.user) {
        const u = tg.initDataUnsafe.user;
        const data = await api('/auth', 'POST', {
            id: u.id, username: u.username || '', first_name: u.first_name
        });
        if (data) {
            currentUser = { id: u.id, username: u.username, first_name: u.first_name };
            loadProfile();
        }
    } else {
        // Demo mode
        currentUser = { id: 111, username: 'demo', first_name: 'Анна' };
        document.getElementById('profileName').textContent = 'Анна (демо)';
    }
}

// ========== TABS ==========
function initTabs() {
    document.querySelectorAll('.tab').forEach(t => {
        t.addEventListener('click', () => {
            document.querySelectorAll('.tab').forEach(x => x.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(x => x.classList.remove('active'));
            t.classList.add('active');
            document.getElementById('tab-' + t.dataset.tab).classList.add('active');
            if (t.dataset.tab === 'gifts') loadGiftShop();
            if (t.dataset.tab === 'matches') loadMatches();
        });
    });

    document.querySelectorAll('.sub-tab').forEach(t => {
        t.addEventListener('click', () => {
            document.querySelectorAll('.sub-tab').forEach(x => x.classList.remove('active'));
            t.classList.add('active');
            const type = t.dataset.subtab;
            document.getElementById('giftShop').style.display = type === 'shop' ? 'grid' : 'none';
            document.getElementById('giftList').style.display = type !== 'shop' ? 'block' : 'none';
            if (type === 'received') loadReceivedGifts();
            if (type === 'sent') loadSentGifts();
        });
    });
}

// ========== SWIPE ==========
function initSwipe() {
    document.getElementById('btnLike').addEventListener('click', () => swipe('like'));
    document.getElementById('btnSkip').addEventListener('click', () => swipe('skip'));
}

async function loadProfiles() {
    const data = await api('/profiles');
    if (data && data.length > 0) {
        profiles = data;
        currentIndex = 0;
        document.getElementById('emptyState').style.display = 'none';
        document.getElementById('swipeActions').style.display = 'flex';
        renderCard();
    } else {
        document.getElementById('emptyState').style.display = 'block';
        document.getElementById('swipeActions').style.display = 'none';
    }
}

function renderCard() {
    if (currentIndex >= profiles.length) {
        document.getElementById('emptyState').style.display = 'block';
        document.getElementById('swipeActions').style.display = 'none';
        return;
    }
    const p = profiles[currentIndex];
    const stack = document.getElementById('cardStack');
    stack.innerHTML = '';
    const card = document.createElement('div');
    card.className = 'profile-card';
    card.innerHTML = `
        <div class="profile-avatar">${p.first_name[0] || '👤'}</div>
        <h2>${p.first_name}, ${p.age}</h2>
        <p class="age-city">📍 ${p.city || 'Не указан'}</p>
        <p class="bio">${p.bio || '...'}</p>
    `;
    stack.appendChild(card);
}

async function swipe(action) {
    if (currentIndex >= profiles.length) return;
    const p = profiles[currentIndex];
    const card = document.querySelector('.profile-card');
    if (card) card.classList.add(action === 'like' ? 'swipe-right' : 'swipe-left');

    if (action === 'like') {
        const res = await api(`/profiles/${p.id}/like`, 'POST');
        if (res?.match) {
            toast('💕 Взаимная симпатия!');
        }
    } else {
        await api(`/profiles/${p.id}/skip`, 'POST');
    }

    setTimeout(() => {
        currentIndex++;
        if (currentIndex < profiles.length) renderCard();
        else {
            document.getElementById('emptyState').style.display = 'block';
            document.getElementById('swipeActions').style.display = 'none';
        }
    }, 400);
}

// ========== GIFTS ==========
let giftShopItems = [];

async function loadGiftShop() {
    const data = await api('/gifts/shop');
    if (!data) return;
    giftShopItems = data;
    const grid = document.getElementById('giftShop');
    grid.innerHTML = data.map(g => `
        <div class="gift-item" onclick="openGiftModal('${g.type}','${g.name}',${g.price},'${g.emoji}')">
            <span class="emoji">${g.emoji}</span>
            <div class="name">${g.name}</div>
            <div class="price">💎 ${g.price}</div>
        </div>
    `).join('');
}

async function loadReceivedGifts() {
    const data = await api('/gifts/received');
    const list = document.getElementById('giftList');
    if (!data?.length) { list.innerHTML = '<p class="empty-msg">Нет полученных подарков</p>'; return; }
    list.innerHTML = data.map(g => `
        <div class="gift-item-detail">
            <span style="font-size:28px">${getEmoji(g.gift_type)}</span>
            <div class="from">
                <strong>${g.gift_name}</strong> от ${g.from_name || 'Аноним'}
                ${g.message ? `<p style="font-size:11px;color:#9e9ece;margin-top:2px">«${g.message}»</p>` : ''}
            </div>
            <span style="font-size:11px;color:#6b6b9e">${new Date(g.date).toLocaleDateString('ru')}</span>
        </div>
    `).join('');
}

async function loadSentGifts() {
    const data = await api('/gifts/sent');
    const list = document.getElementById('giftList');
    if (!data?.length) { list.innerHTML = '<p class="empty-msg">Нет отправленных подарков</p>'; return; }
    list.innerHTML = data.map(g => `
        <div class="gift-item-detail">
            <span style="font-size:28px">${getEmoji(g.gift_type)}</span>
            <div class="from">
                <strong>${g.gift_name}</strong> → ${g.to_name || 'Пользователь'}
                ${g.message ? `<p style="font-size:11px;color:#9e9ece;margin-top:2px">«${g.message}»</p>` : ''}
            </div>
            <span style="font-size:11px;color:#6b6b9e">${new Date(g.date).toLocaleDateString('ru')}</span>
        </div>
    `).join('');
}

function getEmoji(type) {
    const map = { rose_red:'🌹', rose_pink:'🌸', bouquet:'💐', heart:'❤️', diamond:'💎', crown:'👑', ring:'💍', stars:'✨' };
    return map[type] || '🎁';
}

function openGiftModal(type, name, price, emoji) {
    currentGift = { type, name, price, emoji };
    document.getElementById('giftModal').classList.add('active');
    document.getElementById('giftModalTitle').textContent = `${emoji} ${name}`;
    document.getElementById('giftModalDesc').textContent = `Цена: 💎 ${price}`;
    document.getElementById('giftMessage').value = '';
}

async function sendGift() {
    if (!currentGift) return;
    // In full version: select recipient from match list
    const matches = await api('/matches');
    const toUser = matches?.[0]?.id;
    if (!toUser) { toast('Нет симпатий для отправки подарка'); return; }

    const msg = document.getElementById('giftMessage').value;
    const res = await api('/gifts/send', 'POST', {
        to_user: toUser, gift_type: currentGift.type, gift_name: currentGift.name,
        price: currentGift.price, message: msg
    });
    if (res?.error) { toast(res.error); return; }
    document.getElementById('giftModal').classList.remove('active');
    toast('🎁 Подарок отправлен!');
    loadProfile();
}

function initGifts() {
    document.getElementById('btnSendGift').addEventListener('click', sendGift);
    document.getElementById('btnCancelGift').addEventListener('click', () => document.getElementById('giftModal').classList.remove('active'));
}

// ========== MATCHES ==========
async function loadMatches() {
    const data = await api('/matches');
    const list = document.getElementById('matchList');
    if (!data?.length) { list.innerHTML = '<p class="empty-msg">Пока нет взаимных симпатий 💫</p>'; return; }
    list.innerHTML = data.map(m => `
        <div class="match-item">
            <div class="profile-avatar" style="width:50px;height:50px;font-size:20px">${m.first_name[0] || '👤'}</div>
            <div>
                <strong>${m.first_name}, ${m.age}</strong>
                <p style="font-size:12px;color:#6b6b9e">📍 ${m.city || '—'}</p>
            </div>
            <button onclick="sendGiftTo(${m.id})" style="background:#7c5cfc;color:white;border:none;border-radius:10px;padding:6px 12px;font-size:12px">🎁</button>
        </div>
    `).join('');
}

async function sendGiftTo(userId) {
    currentGift = { type: 'rose_red', name: 'Красная роза', price: 1, emoji: '🌹' };
    const res = await api('/gifts/send', 'POST', {
        to_user: userId, gift_type: 'rose_red', gift_name: 'Красная роза', price: 1, message: ''
    });
    if (res?.success) toast('🌹 Роза отправлена!');
    loadProfile();
}

// ========== PROFILE ==========
async function loadProfile() {
    if (!currentUser) return;
    const data = await api('/profile');
    if (!data) return;
    document.getElementById('profileName').textContent = `${data.first_name}, ${data.age}`;
    document.getElementById('profileBio').textContent = data.bio || '...';
    document.getElementById('profileAvatar').textContent = data.first_name?.[0] || '👤';
    document.getElementById('balanceDisplay').textContent = `💎 ${data.balance}`;
}

function initProfile() {
    document.getElementById('btnEditProfile').addEventListener('click', async () => {
        const data = await api('/profile');
        if (!data) return;
        document.getElementById('editName').value = data.first_name;
        document.getElementById('editAge').value = data.age;
        document.getElementById('editCity').value = data.city;
        document.getElementById('editBio').value = data.bio;
        document.getElementById('editProfileModal').classList.add('active');
    });
    document.getElementById('btnSaveProfile').addEventListener('click', async () => {
        const body = {
            first_name: document.getElementById('editName').value,
            age: parseInt(document.getElementById('editAge').value) || 18,
            city: document.getElementById('editCity').value,
            bio: document.getElementById('editBio').value,
        };
        await api('/profile', 'PUT', body);
        document.getElementById('editProfileModal').classList.remove('active');
        loadProfile();
        toast('✅ Профиль обновлён');
    });
    document.getElementById('btnCancelEdit').addEventListener('click', () => {
        document.getElementById('editProfileModal').classList.remove('active');
    });
}

// ========== NOTIFICATIONS ==========
function initNotifications() {
    document.getElementById('btnNotifications').addEventListener('click', async () => {
        const panel = document.getElementById('notifPanel');
        panel.style.display = panel.style.display === 'none' ? 'block' : 'none';
        if (panel.style.display === 'block') await loadNotifications();
    });
    document.getElementById('btnCloseNotifs').addEventListener('click', () => {
        document.getElementById('notifPanel').style.display = 'none';
    });
}

async function loadNotifications() {
    const data = await api('/notifications');
    const list = document.getElementById('notifList');
    if (!data?.length) { list.innerHTML = '<p class="empty-msg">Нет уведомлений</p>'; return; }
    document.getElementById('notifBadge').textContent = data.filter(n => !n.is_read).length;
    list.innerHTML = data.map(n => `
        <div class="notif-item">
            <div class="type">${n.msg_type === 'match' ? '💕 Симпатия' : '🎁 Подарок'}</div>
            <div class="msg">${n.message}</div>
            <div class="time">${new Date(n.date).toLocaleString('ru')}</div>
        </div>
    `).join('');
}

// Periodic polling
setInterval(async () => {
    if (document.getElementById('notifPanel').style.display === 'block') {
        await loadNotifications();
    }
    const data = await api('/notifications');
    if (data) document.getElementById('notifBadge').textContent = data.filter(n => !n.is_read).length || '';
}, 15000);
