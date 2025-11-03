# Ubuntu Server Deployment Guide

Complete guide for deploying the Taxi Service backend on Ubuntu Server.

## Prerequisites

- Ubuntu Server 20.04 LTS or higher
- Root or sudo access
- Domain name (optional but recommended)
- At least 1GB RAM, 10GB disk space

## Table of Contents

1. [Server Setup](#server-setup)
2. [Install Dependencies](#install-dependencies)
3. [PostgreSQL Setup](#postgresql-setup)
4. [Application Deployment](#application-deployment)
5. [Systemd Service Configuration](#systemd-service-configuration)
6. [Nginx Configuration](#nginx-configuration)
7. [SSL/TLS Configuration](#ssltls-configuration)
8. [Monitoring and Logs](#monitoring-and-logs)
9. [Backup Strategy](#backup-strategy)
10. [Troubleshooting](#troubleshooting)

---

## Server Setup

### 1. Update System

```bash
sudo apt update
sudo apt upgrade -y
```

### 2. Create Application User

```bash
sudo adduser --system --group --no-create-home taxi
```

### 3. Configure Firewall

```bash
sudo ufw allow 22/tcp      # SSH
sudo ufw allow 80/tcp      # HTTP
sudo ufw allow 443/tcp     # HTTPS
sudo ufw enable
```

---

## Install Dependencies

### 1. Install Go

```bash
# Download Go (check for latest version at https://go.dev/dl/)
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz

# Extract
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

### 2. Install Git

```bash
sudo apt install git -y
```

### 3. Install Nginx

```bash
sudo apt install nginx -y
sudo systemctl enable nginx
sudo systemctl start nginx
```

---

## PostgreSQL Setup

### 1. Install PostgreSQL

```bash
sudo apt install postgresql postgresql-contrib -y
```

### 2. Configure PostgreSQL

```bash
# Switch to postgres user
sudo -i -u postgres

# Create database
createdb taxi_service

# Create user and grant privileges
psql
```

In PostgreSQL prompt:

```sql
CREATE USER taxi_user WITH PASSWORD 'your_secure_password_here';
ALTER USER taxi_user WITH SUPERUSER;
GRANT ALL PRIVILEGES ON DATABASE taxi_service TO taxi_user;
\q
```

Exit postgres user:

```bash
exit
```

### 3. Configure PostgreSQL for Remote Access (if needed)

Edit PostgreSQL configuration:

```bash
sudo nano /etc/postgresql/*/main/postgresql.conf
```

Find and uncomment:

```
listen_addresses = 'localhost'
```

Edit pg_hba.conf:

```bash
sudo nano /etc/postgresql/*/main/pg_hba.conf
```

Add:

```
host    taxi_service    taxi_user    127.0.0.1/32    md5
```

Restart PostgreSQL:

```bash
sudo systemctl restart postgresql
```

---

## Application Deployment

### 1. Create Application Directory

```bash
sudo mkdir -p /opt/taxi-service
sudo chown taxi:taxi /opt/taxi-service
```

### 2. Clone Repository

```bash
cd /opt/taxi-service
sudo -u taxi git clone <your-repository-url> .
```

Or upload files via SCP:

```bash
# From your local machine
scp -r /path/to/TAXI user@your-server:/tmp/

# On server
sudo mv /tmp/TAXI/* /opt/taxi-service/
sudo chown -R taxi:taxi /opt/taxi-service
```

### 3. Configure Environment

```bash
cd /opt/taxi-service
sudo -u taxi nano .env
```

Add configuration:

```env
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=127.0.0.1
ENV=production

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=taxi_user
DB_PASSWORD=your_secure_password_here
DB_NAME=taxi_service
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your_very_long_random_jwt_secret_key_here
JWT_EXPIRATION_HOURS=720

# File Upload Configuration
UPLOAD_DIR=/opt/taxi-service/uploads
MAX_UPLOAD_SIZE=10485760

# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
TELEGRAM_ADMIN_GROUP_ID=your_admin_group_id

# CORS Configuration
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

**Security Note**: Generate a strong JWT secret:

```bash
openssl rand -base64 64
```

### 4. Create Upload Directory

```bash
sudo -u taxi mkdir -p /opt/taxi-service/uploads
sudo -u taxi mkdir -p /opt/taxi-service/uploads/avatars
sudo -u taxi mkdir -p /opt/taxi-service/uploads/licenses
```

### 5. Build Application

```bash
cd /opt/taxi-service
sudo -u taxi /usr/local/go/bin/go mod download
sudo -u taxi /usr/local/go/bin/go build -o taxi-service cmd/main.go
```

### 6. Test Application

```bash
sudo -u taxi ./taxi-service
```

If successful, you should see:

```
Database connected successfully
Database schema initialized successfully
Starting server on 127.0.0.1:8080
```

Press Ctrl+C to stop and proceed with systemd setup.

---

## Systemd Service Configuration

### 1. Create Service File

```bash
sudo nano /etc/systemd/system/taxi-service.service
```

Add the following:

```ini
[Unit]
Description=Taxi Service Backend API
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=taxi
Group=taxi
WorkingDirectory=/opt/taxi-service
ExecStart=/opt/taxi-service/taxi-service
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=taxi-service

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/taxi-service/uploads

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
```

### 2. Enable and Start Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable taxi-service

# Start service
sudo systemctl start taxi-service

# Check status
sudo systemctl status taxi-service
```

### 3. Verify Service

```bash
# Check if service is running
sudo systemctl is-active taxi-service

# View logs
sudo journalctl -u taxi-service -f
```

---

## Nginx Configuration

### 1. Create Nginx Configuration

```bash
sudo nano /etc/nginx/sites-available/taxi-service
```

Add the following:

```nginx
upstream taxi_backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Logging
    access_log /var/log/nginx/taxi-service-access.log;
    error_log /var/log/nginx/taxi-service-error.log;

    # File upload size limit
    client_max_body_size 10M;

    # Serve static files (uploads)
    location /uploads/ {
        alias /opt/taxi-service/uploads/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # API endpoints
    location /api/ {
        proxy_pass http://taxi_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Swagger documentation
    location /swagger/ {
        proxy_pass http://taxi_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Health check
    location /health {
        proxy_pass http://taxi_backend;
        access_log off;
    }

    # Root redirect
    location / {
        return 301 /swagger/index.html;
    }
}
```

### 2. Enable Configuration

```bash
# Create symbolic link
sudo ln -s /etc/nginx/sites-available/taxi-service /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

---

## SSL/TLS Configuration

### Using Let's Encrypt (Recommended)

#### 1. Install Certbot

```bash
sudo apt install certbot python3-certbot-nginx -y
```

#### 2. Obtain Certificate

```bash
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
```

Follow the prompts. Certbot will automatically:
- Obtain SSL certificate
- Configure Nginx
- Set up automatic renewal

#### 3. Verify Auto-Renewal

```bash
sudo certbot renew --dry-run
```

#### 4. Updated Nginx Configuration

After certbot, your Nginx config will include:

```nginx
server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Rest of configuration...
}

server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;
    return 301 https://$server_name$request_uri;
}
```

---

## Monitoring and Logs

### View Application Logs

```bash
# Real-time logs
sudo journalctl -u taxi-service -f

# Last 100 lines
sudo journalctl -u taxi-service -n 100

# Logs from today
sudo journalctl -u taxi-service --since today

# Logs from specific time
sudo journalctl -u taxi-service --since "2025-11-03 10:00:00"
```

### View Nginx Logs

```bash
# Access logs
sudo tail -f /var/log/nginx/taxi-service-access.log

# Error logs
sudo tail -f /var/log/nginx/taxi-service-error.log
```

### System Resource Monitoring

```bash
# Check service status
sudo systemctl status taxi-service

# Check memory usage
free -h

# Check disk usage
df -h

# Check process
ps aux | grep taxi-service

# Monitor in real-time
htop
```

### Log Rotation

Create log rotation configuration:

```bash
sudo nano /etc/logrotate.d/taxi-service
```

Add:

```
/var/log/nginx/taxi-service-*.log {
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 www-data adm
    sharedscripts
    postrotate
        if [ -f /var/run/nginx.pid ]; then
            kill -USR1 `cat /var/run/nginx.pid`
        fi
    endscript
}
```

---

## Backup Strategy

### 1. Database Backup Script

Create backup script:

```bash
sudo nano /usr/local/bin/backup-taxi-db.sh
```

Add:

```bash
#!/bin/bash

# Configuration
BACKUP_DIR="/opt/backups/taxi-service"
DB_NAME="taxi_service"
DB_USER="taxi_user"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/taxi_db_$TIMESTAMP.sql.gz"
DAYS_TO_KEEP=7

# Create backup directory
mkdir -p $BACKUP_DIR

# Perform backup
PGPASSWORD='your_secure_password_here' pg_dump -U $DB_USER -h localhost $DB_NAME | gzip > $BACKUP_FILE

# Remove old backups
find $BACKUP_DIR -name "taxi_db_*.sql.gz" -mtime +$DAYS_TO_KEEP -delete

echo "Backup completed: $BACKUP_FILE"
```

Make executable:

```bash
sudo chmod +x /usr/local/bin/backup-taxi-db.sh
```

### 2. Schedule Daily Backups

```bash
sudo crontab -e
```

Add:

```cron
# Daily database backup at 2 AM
0 2 * * * /usr/local/bin/backup-taxi-db.sh >> /var/log/taxi-backup.log 2>&1
```

### 3. Backup Uploads

```bash
# Manual backup
sudo tar -czf /opt/backups/taxi-uploads-$(date +%Y%m%d).tar.gz -C /opt/taxi-service uploads/

# Automated daily backup (add to crontab)
0 3 * * * tar -czf /opt/backups/taxi-uploads-$(date +\%Y\%m\%d).tar.gz -C /opt/taxi-service uploads/
```

### 4. Restore Database

```bash
# Restore from backup
gunzip < /opt/backups/taxi-service/taxi_db_20251103_020000.sql.gz | \
PGPASSWORD='your_secure_password_here' psql -U taxi_user -h localhost taxi_service
```

---

## Application Updates

### 1. Update Application

```bash
cd /opt/taxi-service

# Stop service
sudo systemctl stop taxi-service

# Backup current version
sudo -u taxi cp taxi-service taxi-service.backup

# Pull latest changes
sudo -u taxi git pull

# Or upload new files via SCP

# Rebuild
sudo -u taxi /usr/local/go/bin/go build -o taxi-service cmd/main.go

# Start service
sudo systemctl start taxi-service

# Check status
sudo systemctl status taxi-service
```

### 2. Rollback if Issues

```bash
sudo systemctl stop taxi-service
sudo -u taxi cp taxi-service.backup taxi-service
sudo systemctl start taxi-service
```

---

## Security Best Practices

### 1. Firewall Configuration

```bash
# Only allow necessary ports
sudo ufw status
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 2. Fail2Ban (Optional)

Protect against brute force attacks:

```bash
sudo apt install fail2ban -y
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

### 3. Regular Updates

```bash
# Weekly security updates
sudo apt update
sudo apt upgrade -y
sudo apt autoremove -y
```

### 4. Change Default Credentials

```bash
# After first deployment, login as superadmin and:
# 1. Change password via API
# 2. Create your actual admin accounts
# 3. Consider disabling or deleting the default superadmin
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check status
sudo systemctl status taxi-service

# Check logs
sudo journalctl -u taxi-service -n 50

# Common issues:
# 1. Database connection - check credentials in .env
# 2. Port already in use - check with: sudo lsof -i :8080
# 3. Permission issues - check file ownership: ls -la /opt/taxi-service
```

### Database Connection Issues

```bash
# Test database connection
sudo -u taxi psql -h localhost -U taxi_user -d taxi_service -c "\dt"

# Check PostgreSQL status
sudo systemctl status postgresql

# Check PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-*-main.log
```

### Nginx Issues

```bash
# Test configuration
sudo nginx -t

# Check nginx status
sudo systemctl status nginx

# Check nginx logs
sudo tail -f /var/log/nginx/error.log
```

### High Memory Usage

```bash
# Check memory
free -h

# Restart service to free memory
sudo systemctl restart taxi-service

# Check for memory leaks in logs
sudo journalctl -u taxi-service | grep -i "memory\|panic\|fatal"
```

### Disk Space Issues

```bash
# Check disk usage
df -h

# Find large files
sudo du -sh /opt/taxi-service/* | sort -rh | head -10

# Clean old logs
sudo journalctl --vacuum-time=7d

# Clean old backups
find /opt/backups -mtime +30 -delete
```

---

## Performance Optimization

### 1. PostgreSQL Tuning

Edit PostgreSQL configuration:

```bash
sudo nano /etc/postgresql/*/main/postgresql.conf
```

Optimize for your server (example for 2GB RAM):

```
shared_buffers = 512MB
effective_cache_size = 1536MB
maintenance_work_mem = 128MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 8MB
min_wal_size = 1GB
max_wal_size = 4GB
```

Restart PostgreSQL:

```bash
sudo systemctl restart postgresql
```

### 2. Nginx Caching

Add to Nginx configuration:

```nginx
# Cache configuration
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=taxi_cache:10m max_size=100m inactive=60m;
proxy_cache_key "$scheme$request_method$host$request_uri";

# In location block:
location /api/ {
    proxy_cache taxi_cache;
    proxy_cache_valid 200 5m;
    proxy_cache_bypass $http_pragma $http_authorization;
    # ... rest of config
}
```

---

## Monitoring Setup (Optional)

### Using Prometheus + Grafana

This is optional but recommended for production monitoring.

See official documentation:
- [Prometheus](https://prometheus.io/docs/introduction/overview/)
- [Grafana](https://grafana.com/docs/)

---

## Support and Maintenance

### Health Check Endpoint

```bash
# Check if service is healthy
curl http://localhost:8080/health

# Should return: {"status":"ok"}
```

### API Documentation Access

After deployment, access Swagger documentation at:

```
https://yourdomain.com/swagger/index.html
```

---

## Checklist for Production Deployment

- [ ] Server setup and hardening
- [ ] PostgreSQL installed and configured
- [ ] Application built and deployed
- [ ] Environment variables configured
- [ ] Systemd service created and enabled
- [ ] Nginx configured as reverse proxy
- [ ] SSL/TLS certificate installed
- [ ] Firewall configured
- [ ] Backup script configured and tested
- [ ] Log rotation configured
- [ ] Default superadmin password changed
- [ ] Admin accounts created
- [ ] API documentation accessible
- [ ] Health check endpoint responding
- [ ] Test orders created and processed
- [ ] Monitoring configured (optional)

---

## Getting Help

If you encounter issues:

1. Check logs: `sudo journalctl -u taxi-service -f`
2. Verify configuration: `.env` file settings
3. Test database connection
4. Check Nginx configuration: `sudo nginx -t`
5. Review this guide's troubleshooting section

For additional support, refer to the main [README.md](README.md) or create an issue in the repository.

---

**Congratulations!** Your Taxi Service backend is now deployed and ready for production use.
