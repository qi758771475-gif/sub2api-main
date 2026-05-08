# Sub2API

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.25.7-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D.svg)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**AI API Gateway Platform for Subscription Quota Distribution**

</div>

---

## Overview

Sub2API is an AI API gateway platform designed to distribute and manage API quotas from AI product subscriptions. Users can access upstream AI services through platform-generated API Keys, while the platform handles authentication, billing, load balancing, and request forwarding.

## Features

- **Multi-Account Management** - Support multiple upstream account types (OAuth, API Key)
- **API Key Distribution** - Generate and manage API Keys for users
- **Precise Billing** - Token-level usage tracking and cost calculation
- **Smart Scheduling** - Intelligent account selection with sticky sessions
- **Concurrency Control** - Per-user and per-account concurrency limits
- **Rate Limiting** - Configurable request and token rate limits
- **Built-in Payment System** - Supports EasyPay, Alipay, WeChat Pay, and Stripe
- **Admin Dashboard** - Web interface for monitoring and management
- **External System Integration** - Embed external systems via iframe

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.25.7, Gin, Ent |
| Frontend | Vue 3.4+, Vite 5+, TailwindCSS |
| Database | PostgreSQL 15+ |
| Cache/Queue | Redis 7+ |

---

## Docker Deployment (Recommended)

### Prerequisites

- Docker 20.10+
- Docker Compose v2+

### One-Click Deployment

```bash
# Create deployment directory
mkdir -p /opt/sub2api && cd /opt/sub2api

# Download and run deployment preparation script
curl -sSL https://raw.githubusercontent.com/qi758771475-gif/sub2api-main/master/deploy/docker-deploy.sh | bash

# Start services
docker compose up -d

# View logs
docker compose logs -f sub2api
```

**What the script does:**
- Downloads `docker-compose.yml` and `.env.example`
- Generates secure credentials (JWT_SECRET, TOTP_ENCRYPTION_KEY, POSTGRES_PASSWORD)
- Creates `.env` file with auto-generated secrets
- Creates data directories for easy backup/migration

### Manual Deployment

```bash
# 1. Clone the repository
git clone https://github.com/qi758771475-gif/sub2api-main.git
cd sub2api-main/deploy

# 2. Copy environment configuration
cp .env.example .env

# 3. Edit configuration
nano .env

# 4. Create data directories
mkdir -p data postgres_data redis_data

# 5. Start services
docker compose up -d
```

### Access

Open `http://YOUR_SERVER_IP:8080` in your browser.

If admin password was auto-generated, find it in logs:
```bash
docker compose logs sub2api | grep -i password
```

### Upgrade

```bash
docker compose pull
docker compose up -d
```

### Migration

```bash
# Stop services
docker compose down

# Backup entire directory
tar czf sub2api-backup.tar.gz /opt/sub2api/

# Transfer and restore on new server
tar xzf sub2api-backup.tar.gz -C /
cd /opt/sub2api
docker compose up -d
```

---

## Binary Installation

### Prerequisites

- Linux server (amd64 or arm64)
- PostgreSQL 15+
- Redis 7+
- Root privileges

### Install

```bash
curl -sSL https://raw.githubusercontent.com/qi758771475-gif/sub2api-main/master/deploy/install.sh | sudo bash
```

### Post-Installation

```bash
sudo systemctl start sub2api
sudo systemctl enable sub2api
# Open http://YOUR_SERVER_IP:8080
```

### Uninstall

```bash
curl -sSL https://raw.githubusercontent.com/qi758771475-gif/sub2api-main/master/deploy/install.sh | sudo bash -s -- uninstall -y
```

---

## Build from Source

### Prerequisites

- Go 1.21+
- Node.js 18+
- pnpm
- PostgreSQL 15+
- Redis 7+

### Build

```bash
# Clone
git clone https://github.com/qi758771475-gif/sub2api-main.git
cd sub2api-main

# Build frontend
cd frontend
pnpm install
pnpm run build

# Build backend
cd ../backend
go build -tags embed -o sub2api ./cmd/server

# Configure
cp ../deploy/config.example.yaml ./config.yaml
nano config.yaml

# Run
./sub2api
```

---

## Simple Mode

For individual/internal use without SaaS features:

```bash
RUN_MODE=simple
```

---

## Project Structure

```
sub2api/
├── backend/                  # Go backend service
│   ├── cmd/server/           # Application entry
│   ├── internal/             # Internal modules
│   └── resources/            # Static resources
├── frontend/                 # Vue 3 frontend
│   └── src/
└── deploy/                   # Deployment files
    ├── docker-compose.yml
    ├── docker-compose.local.yml
    ├── .env.example
    ├── docker-deploy.sh
    └── install.sh
```

## License

This project is licensed under the [GNU Lesser General Public License v3.0](LICENSE).
