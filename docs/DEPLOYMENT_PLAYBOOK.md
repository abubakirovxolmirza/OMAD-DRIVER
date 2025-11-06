# Deployment Playbook

This document condenses the full `DEPLOYMENT.md` guide into a fast, repeatable runbook for shipping the Taxi Service backend to a fresh Ubuntu host.

| Phase | Goal | Key Commands |
|-------|------|--------------|
| 1. Bootstrap | Harden OS, install tooling | `apt update && apt upgrade -y` · `adduser --system --group taxi` · `ufw allow 22 80 443 && ufw enable` |
| 2. Install Stack | Go, Git, Nginx, PostgreSQL | `wget https://go.dev/dl/go1.21.x.linux-amd64.tar.gz` · `apt install git nginx postgresql postgresql-contrib -y` |
| 3. Database | Create DB + user | `sudo -u postgres createdb taxi_service` · `sudo -u postgres psql` → `CREATE USER ...; GRANT ...;` |
| 4. App Files | Put code in place | `mkdir -p /opt/taxi-service` · `rsync` or `git clone` |
| 5. Config | Provide `.env` | copy `.env.example` → `.env`; fill `SERVER_DOMAIN`, DB creds, JWT secret, CORS origins |
| 6. Build | Get binary + dependencies | `go mod tidy` · `go build -o taxi-service cmd/main.go` |
| 7. Seed Data | Load regions/pricing | `make db-seed` (runs `cmd/tools/dbseed`) |
| 8. Service | Systemd unit for app | `/etc/systemd/system/taxi-service.service`; `systemctl daemon-reload && systemctl enable --now taxi-service` |
| 9. Proxy | Nginx reverse proxy + TLS | `/etc/nginx/sites-available/taxi-service`; `nginx -t`; `certbot --nginx -d api.omad-driver.uz` |
| 10. Smoke Test | Verify endpoints | `curl https://api.omad-driver.uz/health`; hit swagger |
| 11. Ops | Monitoring & backups | `journalctl -u taxi-service -f`; schedule DB backups; verify log rotation |

---

## Detailed Steps

### 1. Bootstrap the server

```bash
sudo apt update && sudo apt upgrade -y
sudo adduser --system --group --no-create-home taxi
sudo ufw allow 22/tcp 80/tcp 443/tcp
sudo ufw --force enable
```

### 2. Install required software

```bash
# Go (replace with latest minor version)
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh
source /etc/profile.d/go.sh

sudo apt install git nginx postgresql postgresql-contrib -y
```

### 3. Prepare PostgreSQL

```bash
sudo -u postgres createdb taxi_service
sudo -u postgres psql
```

Inside `psql`:

```sql
CREATE USER taxi_user WITH PASSWORD 'superStrongPassword!';
GRANT ALL PRIVILEGES ON DATABASE taxi_service TO taxi_user;
```

For remote connections, update `postgresql.conf` (`listen_addresses`) and `pg_hba.conf`, then restart:

```bash
sudo systemctl restart postgresql
```

### 4. Deploy application files

```bash
sudo mkdir -p /opt/taxi-service
sudo chown taxi:taxi /opt/taxi-service
cd /opt/taxi-service
sudo -u taxi git clone <repo-url> .
# or copy build artifacts via scp/rsync
```

### 5. Configure environment

```bash
sudo -u taxi cp .env.example .env
sudo -u taxi nano .env
```

Set at minimum:

```
SERVER_PORT=8080
SERVER_HOST=127.0.0.1
SERVER_DOMAIN=api.omad-driver.uz
DB_HOST=localhost
DB_PORT=5432
DB_USER=taxi_user
DB_PASSWORD=superStrongPassword!
DB_NAME=taxi_service
JWT_SECRET=<generate with openssl rand -base64 64>
CORS_ALLOWED_ORIGINS=https://api.omad-driver.uz,https://docs.omad-driver.uz
UPLOAD_DIR=/opt/taxi-service/uploads
```

### 6. Build & install dependencies

```bash
sudo -u taxi /usr/local/go/bin/go mod tidy
sudo -u taxi /usr/local/go/bin/go build -o taxi-service cmd/main.go
sudo -u taxi mkdir -p uploads/avatars uploads/licenses
```

### 7. Seed reference data (optional but recommended)

```bash
sudo -u taxi make db-seed      # loads regions, districts, pricing, discounts
```

### 8. Configure systemd service

Create `/etc/systemd/system/taxi-service.service`:

```ini
[Unit]
Description=Taxi Service Backend API
After=network.target postgresql.service

[Service]
User=taxi
Group=taxi
WorkingDirectory=/opt/taxi-service
ExecStart=/opt/taxi-service/taxi-service
Restart=always
RestartSec=5
EnvironmentFile=/opt/taxi-service/.env
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable taxi-service
sudo systemctl start taxi-service
sudo systemctl status taxi-service
```

### 9. Configure Nginx + HTTPS

Create `/etc/nginx/sites-available/taxi-service`:

```nginx
upstream taxi_backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name api.omad-driver.uz;

    client_max_body_size 10M;

    location /uploads/ {
        alias /opt/taxi-service/uploads/;
        expires 30d;
    }

    location / {
        proxy_pass http://taxi_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable & reload:

```bash
sudo ln -s /etc/nginx/sites-available/taxi-service /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

Install TLS via certbot:

```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d api.omad-driver.uz -d docs.omad-driver.uz
```

### 10. Smoke tests

```bash
curl https://api.omad-driver.uz/health
# -> {"status":"ok"}

# JWT login to verify role payload
curl -X POST https://api.omad-driver.uz/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"phone_number":"...","password":"..."}'

# Swagger UI
open https://api.omad-driver.uz/swagger/index.html
```

### 11. Ongoing operations

- Logs: `sudo journalctl -u taxi-service -f`
- Restart after deploy: `sudo systemctl restart taxi-service`
- Backups: use the script and cron job described in `DEPLOYMENT.md`
- Security: keep UFW strict, rotate JWT secret, disable default superadmin credentials in production.

---

## Quick Checklist

- [ ] OS updated, firewall enabled
- [ ] Go/Git/Nginx/PostgreSQL installed
- [ ] `taxi_service` DB + dedicated user created
- [ ] Code deployed under `/opt/taxi-service`
- [ ] `.env` configured with domain, DB creds, JWT secret
- [ ] Binary built and reference data seeded
- [ ] Systemd service runs and survives reboot
- [ ] Nginx reverse proxy + TLS certificate active
- [ ] `/health` and `/swagger` responding over HTTPS
- [ ] Backups & monitoring set up

Refer back to `DEPLOYMENT.md` for troubleshooting, performance tuning, and advanced hardening once this playbook is complete.
