# 🚀 StdRequestBot — Telegram Automated Join Request Approver

<p align="center">
  <img src="https://graph.org/file/00ea4effe5d2dfbb8d5be.jpg" alt="StdRequestBot Banner" width="450"/>
</p>

<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Language-Go%201.22+-00ADD8?style=for-the-badge&logo=go" alt="Go"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-AGPL%20v3-blue?style=for-the-badge" alt="License"/></a>
  <a href="https://t.me/STDBOTS"><img src="https://img.shields.io/badge/Channel-%40STDBOTS-2CA5E0?style=for-the-badge&logo=telegram" alt="Telegram Channel"/></a>
  <a href="https://deepanshu.in"><img src="https://img.shields.io/badge/Author-STD%20DEEPANSHU-FF4500?style=for-the-badge" alt="Author"/></a>
</p>

<p align="center">
  <a href="https://dashboard.heroku.com/new?template=https://github.com/StdBots/StdRequestBot">
    <img src="https://img.shields.io/badge/Deploy%20To%20Heroku-7056bf?style=for-the-badge&logo=heroku" alt="Deploy to Heroku"/>
  </a>
  <a href="https://railway.app/template/new?template=https://github.com/StdBots/StdRequestBot">
    <img src="https://img.shields.io/badge/Deploy%20On%20Railway-0B0D0E?style=for-the-badge&logo=railway" alt="Deploy on Railway"/>
  </a>
</p>

---

## ⚡ Overview

**StdRequestBot** is an ultra-fast, high-concurrency Telegram bot engineered in **Go (Golang)** with **MongoDB** persistence to automatically approve channel and group join requests. 

It implements the critical **"DM-First Protocol"**: sending the welcome/marketing message directly to the user's private chat *before* approving the request (leveraging Telegram's temporary DM authorization while a request is pending), thereby building a guaranteed 100% reachable lead generation database for mass broadcasting.

Developed by **[STD DEEPANSHU](https://deepanshu.in)** as part of the **[STD BOTS Ecosystem](https://t.me/STDBOTS)**.

---

## 🔑 The Critical "DM-First" Protocol

Telegram states:
> *"The bot can use this identifier to send messages UNTIL THE JOIN REQUEST IS PROCESSED, assuming no other administrator contacted the user."*

If a bot calls `approveChatJoinRequest()` first, Telegram immediately marks the request as processed and revokes the privilege to send a private message to the user (causing `403 Forbidden: bot can't initiate conversation`).

**StdRequestBot executes the proper pipeline:**
1. **Deduplication Check (180s):** Thread-safe monotonic cache eliminates duplicate processing.
2. **Step 1 (Send DM):** Dispatches customized welcome message with inline buttons to user's PM while request is **PENDING**.
3. **Step 2 (Approve):** Calls `approveChatJoinRequest(chat_id, user_id)`.
4. **Step 3 (Lead Capture):** Only if the DM succeeded, saves user in MongoDB for future `/broadcast` marketing!

---

## ✨ Features

- ⚡ **Go 1.22+ Concurrency:** Goroutines process thousands of join requests per minute without lag.
- 📬 **DM-First Funnel:** Guaranteed welcome message delivery before approval.
- 🏢 **Multi-Channel Management:** Connect and manage unlimited channels and groups simultaneously.
- 🎨 **Custom Welcome Templates:** Configure unique welcome messages and buttons per channel (`/setwelcome`).
- ⏱️ **Instant or Delayed Approval:** Support for 0-second instant approval or anti-flood staggered delays.
- 👥 **Mass Lead Generation Database:** Automatically builds an audience of reachable users in MongoDB.
- 📢 **Admin Broadcast Engine:** High-speed bulk broadcasting with worker pools and flood wait protection.
- 🔒 **7-Layer Credit Protection:** AGPL-3.0 integrity validation, zero-width watermarks, and brand protection.

---

## 🚀 One-Click Deployments

### 🟣 Deploy to Heroku
Click the button below to deploy your instance to Heroku in 60 seconds:

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://dashboard.heroku.com/new?template=https://github.com/StdBots/StdRequestBot)

### 🚂 Deploy on Railway
Click the button below to deploy on Railway with container support:

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/template/new?template=https://github.com/StdBots/StdRequestBot)

### 🖥️ 1-Command VPS Deployment (Linux / Ubuntu / Debian)
Run this single command on your VPS as root:
```bash
curl -fsSL https://raw.githubusercontent.com/StdBots/StdRequestBot/main/scripts/install_vps.sh | bash
```

---

## 🐳 Docker & Manual VPS Setup

```bash
# 1. Clone repository
git clone https://github.com/StdBots/StdRequestBot.git
cd StdRequestBot

# 2. Configure environment
cp .env.example .env
nano .env

# 3. Start with Docker Compose
docker compose up -d --build
```

---

## ⚙️ Environment Variables

| Variable | Description | Required | Default |
|---|---|---|---|
| `BOT_TOKEN` | Telegram Bot Token from [@BotFather](https://t.me/BotFather) | **Yes** | — |
| `OWNER_ID` | Telegram User ID of the primary administrator | **Yes** | `7394590844` |
| `MONGO_URI` | MongoDB Connection String (Atlas or Local) | **Yes** | `mongodb://localhost:27017/stdrequestbot` |
| `DB_NAME` | Database name | No | `stdrequestbot` |
| `FORCE_SUB_CHANNEL` | Channel username without `@` for force-sub | No | `StdBots` |
| `LOG_CHANNEL_ID` | Telegram Channel ID for logging events | No | `0` |
| `APPROVAL_DELAY_SECONDS` | Delay before approving (0 = instant) | No | `0` |
| `ENV` | Environment mode (`development`/`production`) | No | `production` |

---

## 🤖 Commands

| Command | Description |
|---|---|
| `/start` | Launch bot, view features and credits |
| `/channels` | View all connected channels and approved counts |
| `/setwelcome` | Set custom welcome message for a channel |
| `/delwelcome` | Reset to default welcome template |
| `/help` | Detailed help and setup guide |
| `/stats` | Global approval statistics & system metrics (Admin only) |
| `/broadcast` | Broadcast message to all registered users (Admin only) |

---

## 📄 License & Attribution

Licensed under the [GNU Affero General Public License v3 (AGPL-3.0)](LICENSE).

Mandatory Attribution: Derivative works, forks, and hosted instances must preserve all visible and embedded credits pointing to **STD DEEPANSHU** ([https://deepanshu.in](https://deepanshu.in)) and **STD BOTS** ([@STDBOTS](https://t.me/STDBOTS)).
