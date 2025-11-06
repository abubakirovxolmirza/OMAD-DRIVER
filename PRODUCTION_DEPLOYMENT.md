# Production Deployment Guide - Taxi Service API

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Deployment Options](#deployment-options)
3. [Docker Deployment](#docker-deployment)
4. [Systemd Service Setup](#systemd-service-setup)
5. [Nginx Configuration](#nginx-configuration)
6. [SSL/TLS Certificate](#ssltls-certificate)
7. [Database Backup](#database-backup)
8. [Monitoring & Logging](#monitoring--logging)
9. [Troubleshooting](#troubleshooting)

---

## Prerequisites

- Server OS: Ubuntu 20.04 or higher (recommended)
- Docker & Docker Compose (for containerized deployment)
- PostgreSQL 15+ (if not using Docker)
- Nginx (as reverse proxy)
- Domain name (e.g., api.omad-driver.uz)
- SSL/TLS certificate (Let's Encrypt recommended)
- 2+ GB RAM, 20+ GB storage

---

## Deployment Options

### Option 1: Docker Compose (Recommended)
- ✅ Easiest to deploy and scale
- ✅ Self-contained environment
- ✅ Easy rollbacks
- ✅ Perfect for most scenarios

### Option 2: Binary with Systemd
- ✅ Lightweight deployment
- ✅ Direct server control
- ✅ Better for resource-constrained servers

### Option 3: Manual Setup
- ✅ Full control over configuration
- ⚠️ More complex maintenance

---

## Docker Deployment

### Step 1: Prepare Server

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker & Docker Compose
sudo apt install -y docker.io docker-compose curl

# Enable Docker service
sudo systemctl enable docker
sudo systemctl start docker

# Add current user to docker group (optional, restart shell after)
sudo usermod -aG docker $USER
newgrp docker
```

### Step 2: Clone Repository & Configure

```bash
# Clone project
cd /opt
sudo git clone https://github.com/abubakirovxolmirza/OMAD-DRIVER.git
cd OMAD-DRIVER

# Create .env file from example
sudo cp .env.example .env

# Edit configuration (as root or with sudo)
sudo nano .env
```

### Step 3: Configure Environment Variables

Edit `.env` with production values:

```bash
# Security - CHANGE THESE!
JWT_SECRET=$(openssl rand -base64 32)
DB_PASSWORD=<generate_strong_password>

# Domain & CORS
CORS_ALLOWED_ORIGINS=https://api.omad-driver.uz,https://omad-driver.uz

# Email for Telegram notifications (optional)
TELEGRAM_BOT_TOKEN=<your_telegram_bot_token>

# Environment
ENV=production
```

### Step 4: Start Services

```bash
# Build and start
sudo docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f app

# Seed database (optional - adds regions, districts, pricing)
docker-compose exec app /app/cmd/tools/dbseed/main.go -action=seed
```

### Step 5: Verify Deployment

```bash
# Test health endpoint
curl http://localhost:8080/health

# Test API (should get 401 without token)
curl http://localhost:8080/api/v1/regions
```

---

## Systemd Service Setup

### Build Binary

```bash
# On your development machine or CI/CD
go build -o taxi-service cmd/main.go

# Or build on server
cd /opt/OMAD-DRIVER
go build -o taxi-service cmd/main.go
```

### Create Service File

```bash
sudo tee /etc/systemd/system/taxi-service.service > /dev/null << EOF
[Unit]
Description=Taxi Service API
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=taxi
WorkingDirectory=/opt/OMAD-DRIVER
EnvironmentFile=/opt/OMAD-DRIVER/.env
ExecStart=/opt/OMAD-DRIVER/taxi-service
Restart=on-failure
RestartSec=5s

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/opt/OMAD-DRIVER/uploads /var/log/taxi-service

[Install]
WantedBy=multi-user.target
EOF
```

### Enable & Start Service

```bash
# Create user
sudo useradd -r -s /bin/false taxi || true

# Set permissions
sudo chown -R taxi:taxi /opt/OMAD-DRIVER

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable taxi-service
sudo systemctl start taxi-service

# Check status
sudo systemctl status taxi-service

# View logs
sudo journalctl -u taxi-service -f
```

---

## Nginx Configuration

### Install Nginx

```bash
sudo apt install -y nginx
sudo systemctl enable nginx
```

### Create Nginx Configuration

```bash
sudo tee /etc/nginx/sites-available/taxi-api > /dev/null << 'EOF'
# Redirect HTTP to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name api.omad-driver.uz omad-driver.uz;
    
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    
    location / {
        return 301 https://$server_name$request_uri;
    }
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api.omad-driver.uz omad-driver.uz;

    # SSL certificates (update paths as needed)
    ssl_certificate /etc/letsencrypt/live/api.omad-driver.uz/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.omad-driver.uz/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

    # Logging
    access_log /var/log/nginx/taxi-api-access.log combined;
    error_log /var/log/nginx/taxi-api-error.log warn;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml text/javascript application/json application/javascript application/xml+rss;

    # Client max body size (for file uploads)
    client_max_body_size 10M;

    # Proxy settings
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Static files (uploads)
    location /uploads/ {
        alias /opt/OMAD-DRIVER/uploads/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}
EOF
```

### Enable Configuration

```bash
# Create symlink
sudo ln -s /etc/nginx/sites-available/taxi-api /etc/nginx/sites-enabled/

# Remove default config
sudo rm /etc/nginx/sites-enabled/default

# Test configuration
sudo nginx -t

# Restart Nginx
sudo systemctl restart nginx
```

---

## SSL/TLS Certificate

### Using Let's Encrypt (Certbot)

```bash
# Install certbot
sudo apt install -y certbot python3-certbot-nginx

# Obtain certificate
sudo certbot certonly --nginx -d api.omad-driver.uz -d omad-driver.uz

# Auto-renewal (already enabled by default)
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer

# Verify auto-renewal
sudo certbot renew --dry-run
```

---

## Database Backup

### Automated Backup Script

```bash
# Create backup directory
mkdir -p /backups/taxi-db
chmod 700 /backups/taxi-db

# Create backup script
sudo tee /usr/local/bin/backup-taxi-db.sh > /dev/null << 'EOF'
#!/bin/bash
BACKUP_DIR="/backups/taxi-db"
DATE=$(date +%Y-%m-%d_%H-%M-%S)
DB_NAME="taxi_service"
DB_USER="taxi_user"

# Create backup
PGPASSWORD="$DB_PASSWORD" pg_dump -h localhost -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_DIR/backup_$DATE.sql.gz"

# Keep only last 7 days
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +7 -delete

echo "Backup created: $BACKUP_DIR/backup_$DATE.sql.gz"
EOF

chmod +x /usr/local/bin/backup-taxi-db.sh

# Add to crontab (daily at 2 AM)
(crontab -l 2>/dev/null; echo "0 2 * * * /usr/local/bin/backup-taxi-db.sh") | crontab -
```

### Manual Backup

```bash
# Backup database
PGPASSWORD="password" pg_dump -h localhost -U taxi_user taxi_service > backup.sql

# Restore from backup
PGPASSWORD="password" psql -h localhost -U taxi_user taxi_service < backup.sql
```

---

## Monitoring & Logging

### System Monitoring

```bash
# Install htop for resource monitoring
sudo apt install -y htop

# Check service status
sudo systemctl status taxi-service

# Monitor logs
sudo journalctl -u taxi-service -f
```

### Log Rotation

```bash
sudo tee /etc/logrotate.d/taxi-service > /dev/null << 'EOF'
/var/log/taxi-service/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 taxi taxi
    sharedscripts
    postrotate
        systemctl reload taxi-service > /dev/null 2>&1 || true
    endscript
}
EOF
```

### Monitoring with Prometheus (Optional)

```bash
# Add Prometheus endpoint to Nginx config
location /metrics {
    proxy_pass http://localhost:8080/metrics;
}
```

---

## Maintenance Tasks

### Update Application

```bash
# Using Docker
cd /opt/OMAD-DRIVER
git pull origin main
docker-compose up -d --build

# Using Systemd
cd /opt/OMAD-DRIVER
git pull origin main
go build -o taxi-service cmd/main.go
sudo systemctl restart taxi-service
```

### Database Maintenance

```bash
# Connect to database
PGPASSWORD="password" psql -h localhost -U taxi_user taxi_service

# Analyze and vacuum
ANALYZE;
VACUUM;

# Check table sizes
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) 
FROM pg_tables WHERE schemaname != 'pg_catalog' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

---

## Troubleshooting

### Application Won't Start

```bash
# Check logs
docker-compose logs app

# Verify database connection
docker-compose exec app /bin/sh -c 'nc -zv db 5432'

# Check environment variables
docker-compose exec app env | grep DB
```

### Database Connection Issues

```bash
# Check PostgreSQL is running
sudo docker ps | grep postgres

# Check database
docker-compose exec db psql -U taxi_user -d taxi_service -c "\dt"
```

### Port Already in Use

```bash
# Find process using port 8080
sudo lsof -i :8080

# Kill the process
sudo kill -9 <PID>
```

### Nginx Not Proxying

```bash
# Test Nginx config
sudo nginx -t

# Check Nginx logs
sudo tail -f /var/log/nginx/taxi-api-error.log

# Verify app is running on localhost:8080
curl http://localhost:8080/health
```

### SSL Certificate Issues

```bash
# Check certificate validity
echo | openssl s_client -servername api.omad-driver.uz -connect api.omad-driver.uz:443

# Renew certificate
sudo certbot renew --force-renewal

# Check certificate files
ls -la /etc/letsencrypt/live/api.omad-driver.uz/
```

---

## Performance Optimization

### Database Connection Pooling

Connection pooling is already configured in the application. Adjust if needed:

```go
DB.SetMaxOpenConns(25)  // Maximum connections
DB.SetMaxIdleConns(5)   // Idle connections
```

### Caching Headers (in Nginx)

```nginx
# Cache static assets for 30 days
location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg)$ {
    expires 30d;
    add_header Cache-Control "public, immutable";
}
```

### Rate Limiting (Future Enhancement)

```nginx
# Add to Nginx config
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/m;

location /api/v1/ {
    limit_req zone=api_limit burst=20 nodelay;
    proxy_pass http://localhost:8080;
}
```

---

## Rollback Procedure

```bash
# If using Docker
docker-compose down
git checkout <previous_tag>
docker-compose up -d --build

# If using Systemd
sudo systemctl stop taxi-service
git checkout <previous_tag>
go build -o taxi-service cmd/main.go
sudo systemctl start taxi-service
```

---

## Security Checklist

- [ ] Change all default passwords
- [ ] Generate strong JWT secret
- [ ] Enable HTTPS/SSL
- [ ] Configure firewall rules
- [ ] Set up database backups
- [ ] Enable database encryption
- [ ] Configure log rotation
- [ ] Enable audit logging
- [ ] Restrict admin access
- [ ] Update system packages regularly
- [ ] Monitor API usage
- [ ] Set up alerts for errors/failures

---

## Support & Troubleshooting

For deployment issues or questions:
- Email: support@taxiservice.com
- Documentation: See README.md and API_DOCUMENTATION.md
- Bug Reports: Create an issue on GitHub
