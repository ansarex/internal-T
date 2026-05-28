# Deployment Guide — internal-t.com

> Stack: Go API + Astro Frontend + MySQL + DigitalOcean Spaces
> Server: DigitalOcean Droplet (Ubuntu 22.04)

---

## Architecture

```
https://internal-t.com        →  Frontend (Astro SSR, port 4321)
https://api.internal-t.com    →  Backend  (Go API,    port 8080)
```

Nginx sits in front of both and handles SSL via Let's Encrypt.

---

## Step 1 — Create Droplet

1. Go to **DigitalOcean → Droplets → Create**
2. Choose **Ubuntu 22.04**, minimum **2GB RAM / 1 vCPU**
3. Add your SSH key
4. Note the droplet IP address

### Point DNS Records

In your domain registrar (or DigitalOcean DNS), add:

| Type | Name | Value |
|------|------|-------|
| A | `internal-t.com` | `your_droplet_ip` |
| A | `www.internal-t.com` | `your_droplet_ip` |
| A | `api.internal-t.com` | `your_droplet_ip` |

> DNS can take up to 30 minutes to propagate.

---

## Step 2 — Connect & Update Server

```bash
ssh root@your_droplet_ip

apt update && apt upgrade -y
```

---

## Step 3 — Install Dependencies

```bash
# Nginx, MySQL, Certbot, Git
apt install -y nginx mysql-server certbot python3-certbot-nginx git curl ufw

# Go 1.22
wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version   # should print: go version go1.22.3

# Node 20
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt install -y nodejs
node -v      # should print: v20.x.x
```

---

## Step 4 — Configure Firewall

```bash
ufw allow OpenSSH
ufw allow 'Nginx Full'
ufw enable
ufw status
```

---

## Step 5 — Setup MySQL

```bash
mysql_secure_installation
# Follow prompts: set root password, remove anonymous users, disallow remote root login
```

Create the database and user:

```bash
mysql -u root -p
```

```sql
CREATE DATABASE internalt CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'internalt'@'localhost' IDENTIFIED BY 'your_strong_db_password';
GRANT ALL PRIVILEGES ON internalt.* TO 'internalt'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

---

## Step 6 — Clone Repository

```bash
mkdir -p /var/www/internal-t
git clone https://github.com/youruser/internal-t.git /var/www/internal-t
```

---

## Step 7 — Deploy Backend

### 7a. Create production `.env`

```bash
cd /var/www/internal-t/backend
cp .env.example .env
nano .env
```

Fill in all values:

```env
APP_URL=https://api.internal-t.com
FRONTEND_URL=https://internal-t.com
APP_ENV=production
APP_PORT=8080

DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=internalt
DB_USERNAME=internalt
DB_PASSWORD=your_strong_db_password

FILESYSTEM_DISK=s3
AWS_ACCESS_KEY_ID=your_spaces_key
AWS_SECRET_ACCESS_KEY=your_spaces_secret
AWS_DEFAULT_REGION=sgp1
AWS_BUCKET=your-bucket-name
AWS_ENDPOINT=https://sgp1.digitaloceanspaces.com

RESEND_API_KEY=re_xxxxxxxxxxxx
MAIL_FROM_ADDRESS=noreply@internal-t.com
MAIL_FROM_NAME=Trustwired

APP_KEY=run_openssl_rand_base64_32_to_generate_this
```

> Generate APP_KEY: `openssl rand -base64 32`

### 7b. Build & migrate

```bash
cd /var/www/internal-t/backend
go mod tidy
go build -o internal-t-api .
./internal-t-api --migrate
```

### 7c. Seed admin account (first deploy only)

```bash
go run ./cmd/seed/
```

### 7d. Create systemd service

```bash
nano /etc/systemd/system/internal-t-api.service
```

Paste:

```ini
[Unit]
Description=Internal-T API
After=network.target mysql.service

[Service]
Type=simple
User=root
WorkingDirectory=/var/www/internal-t/backend
ExecStart=/var/www/internal-t/backend/internal-t-api
Restart=always
RestartSec=5
EnvironmentFile=/var/www/internal-t/backend/.env

[Install]
WantedBy=multi-user.target
```

```bash
systemctl daemon-reload
systemctl enable internal-t-api
systemctl start internal-t-api
systemctl status internal-t-api
# ✓ should show: Active: active (running)
```

---

## Step 8 — Deploy Frontend

### 8a. Create production `.env`

```bash
cd /var/www/internal-t/frontend
echo "PUBLIC_API_URL=https://api.internal-t.com" > .env
```

### 8b. Build

```bash
npm install
npm run build
```

### 8c. Create systemd service

```bash
nano /etc/systemd/system/internal-t-web.service
```

Paste:

```ini
[Unit]
Description=Internal-T Frontend
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/var/www/internal-t/frontend
ExecStart=/usr/bin/node ./dist/server/entry.mjs
Restart=always
RestartSec=5
Environment=HOST=0.0.0.0
Environment=PORT=4321

[Install]
WantedBy=multi-user.target
```

```bash
systemctl daemon-reload
systemctl enable internal-t-web
systemctl start internal-t-web
systemctl status internal-t-web
# ✓ should show: Active: active (running)
```

---

## Step 9 — Configure Nginx

```bash
nano /etc/nginx/sites-available/internal-t
```

Paste:

```nginx
# Frontend
server {
    server_name internal-t.com www.internal-t.com;

    location / {
        proxy_pass         http://localhost:4321;
        proxy_http_version 1.1;
        proxy_set_header   Upgrade $http_upgrade;
        proxy_set_header   Connection 'upgrade';
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_cache_bypass $http_upgrade;
    }

    listen 80;
}

# Backend API
server {
    server_name api.internal-t.com;

    location / {
        proxy_pass         http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        client_max_body_size 15M;
    }

    listen 80;
}
```

```bash
ln -s /etc/nginx/sites-available/internal-t /etc/nginx/sites-enabled/
nginx -t
# ✓ should print: syntax is ok / test is successful

systemctl reload nginx
```

---

## Step 10 — Enable HTTPS (SSL)

```bash
certbot --nginx -d internal-t.com -d api.internal-t.com
```

Follow the prompts. Certbot will:
- Obtain certificates from Let's Encrypt
- Automatically update your Nginx config for HTTPS
- Set up auto-renewal

Test auto-renewal:

```bash
certbot renew --dry-run
```

---

## Step 11 — Verify Everything

```bash
# Check both services are running
systemctl status internal-t-api
systemctl status internal-t-web

# Test API
curl https://api.internal-t.com/api/me
# Expected: {"message":"Unauthenticated."}  ← means API is reachable

# Test Frontend
curl -I https://internal-t.com
# Expected: HTTP/2 200
```

Open `https://internal-t.com` in your browser — you should see the login page.

Login with:
- **Email:** `admin@test.com`
- **Password:** `password123`

> Change this password immediately after first login.

---

## Redeploying (Future Updates)

```bash
cd /var/www/internal-t
git pull

# Rebuild & restart backend
cd backend
go build -o internal-t-api .
systemctl restart internal-t-api

# Rebuild & restart frontend
cd ../frontend
npm run build
systemctl restart internal-t-web
```

Or save as a script:

```bash
nano /var/www/internal-t/deploy.sh
```

```bash
#!/bin/bash
set -e

echo "==> Pulling latest code..."
cd /var/www/internal-t
git pull

echo "==> Building backend..."
cd backend
go build -o internal-t-api .
systemctl restart internal-t-api

echo "==> Building frontend..."
cd ../frontend
npm run build
systemctl restart internal-t-web

echo "==> Done! Both services restarted."
```

```bash
chmod +x /var/www/internal-t/deploy.sh

# To deploy in future:
bash /var/www/internal-t/deploy.sh
```

---

## Useful Commands

| Command | Description |
|---------|-------------|
| `systemctl status internal-t-api` | Check API status |
| `systemctl status internal-t-web` | Check frontend status |
| `journalctl -u internal-t-api -f` | Live API logs |
| `journalctl -u internal-t-web -f` | Live frontend logs |
| `systemctl restart internal-t-api` | Restart API |
| `systemctl restart internal-t-web` | Restart frontend |
| `nginx -t && systemctl reload nginx` | Reload Nginx config |
| `certbot renew --dry-run` | Test SSL renewal |



--- update databse

⏺ You're using GORM AutoMigrate, so the process depends on what kind of DB change it is:

  ---
  AutoMigrate handles automatically (just deploy normally):
  - New tables
  - New columns
  - New indexes

  AutoMigrate does NOT handle (you must run SQL manually first):
  - Table renames
  - Column renames
  - Dropping columns

  ---
  For this deployment (invoices → receipts rename)

  Since we renamed the table and the invoice_number column, you need to run SQL on prod before deploying the code.
  SSH in and run:

  ssh root@206.189.42.240
  mysql -u root -p internalt

  -- Rename the table
  RENAME TABLE invoices TO receipts;
                                              
  -- Rename the column inside it          
  ALTER TABLE receipts CHANGE invoice_number receipt_number VARCHAR(255) NOT NULL;
                                                                                                                    
  Then deploy normally:                   
                                                                                                                    
  bash /var/www/internal-t/deploy.sh                                                                                
  
  ---                                                                                                               
  For future deployments with DB changes                                                                          
                                              
  Use this rule of thumb:                 

  ┌──────────────────────────────┬─────────────────────────────────────────────────────┐                            
  │         Change type          │                     What to do                      │
  ├──────────────────────────────┼─────────────────────────────────────────────────────┤                            
  │ New table / new column       │ Just deploy — AutoMigrate picks it up on restart    │                            
  ├──────────────────────────────┼─────────────────────────────────────────────────────┤
  │ Rename table / rename column │ Run SQL manually on prod first, then deploy         │                            
  ├──────────────────────────────┼─────────────────────────────────────────────────────┤                          
  │ Drop column                  │ Remove from model first (deploy), then drop via SQL │
  └──────────────────────────────┴─────────────────────────────────────────────────────┘                            
                                          
  For the normal redeploy (no manual SQL needed), your existing deploy.sh script handles everything.                
                                                                                                      