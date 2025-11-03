# Quick Start Guide

Get the Taxi Service backend running in 5 minutes!

## Prerequisites

- Go 1.21+
- PostgreSQL 12+
- Git

## Installation

### 1. Clone & Setup

```bash
# Navigate to project directory
cd /c/Users/Xolmirza/Desktop/TAXI

# Copy environment file
cp .env.example .env

# Edit .env with your database credentials
# Minimum required: DB_PASSWORD, JWT_SECRET
```

### 2. Configure Database

```bash
# Create database
createdb taxi_service

# Or using psql
psql -U postgres
CREATE DATABASE taxi_service;
\q
```

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run Application

```bash
go run cmd/main.go
```

The server will:
- Connect to database ✓
- Create all tables automatically ✓
- Seed initial data (regions, discounts) ✓
- Create default superadmin ✓
- Start on http://localhost:8080 ✓

## Quick Test

### 1. Access Swagger Documentation

Open browser: http://localhost:8080/swagger/index.html

### 2. Login as SuperAdmin

**Default Credentials** (development only):
- Phone: `+998901234567`
- Password: `admin123`

### 3. Test API with curl

```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+998901234567",
    "password": "admin123"
  }'

# Copy the token from response

# Get profile
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## Next Steps

1. **Change Default Password** - Use `/api/v1/auth/change-password`
2. **Create Admin Users** - Use `/api/v1/admin/create-admin`
3. **Set Pricing** - Configure routes via `/api/v1/admin/pricing`
4. **Register Test Users** - Via `/api/v1/auth/register`
5. **Test Driver Application** - Apply as driver and approve

## Common Commands

```bash
# Run application
go run cmd/main.go

# Build binary
go build -o taxi-service cmd/main.go

# Run binary
./taxi-service

# Generate Swagger docs (if you modify API)
swag init -g cmd/main.go -o ./docs

# Run with Make
make run
make build
```

## Troubleshooting

### Database Connection Failed
- Check PostgreSQL is running: `pg_isadmin`
- Verify credentials in `.env`
- Ensure database exists: `psql -l`

### Port Already in Use
```bash
# Find process using port 8080
netstat -ano | findstr :8080  # Windows
lsof -i :8080                  # Linux/Mac

# Kill the process or change SERVER_PORT in .env
```

### Module Errors
```bash
go mod tidy
go mod download
```

## API Endpoints Overview

- **Auth**: `/api/v1/auth/*` - Registration, login, profile
- **Orders**: `/api/v1/orders/*` - Create taxi/delivery orders
- **Driver**: `/api/v1/driver/*` - Driver operations
- **Admin**: `/api/v1/admin/*` - Admin panel
- **Swagger**: `/swagger/index.html` - API documentation

## Default Data

### Regions (Auto-seeded)
- Toshkent, Samarqand, Buxoro, Andijon, Farg'ona, Namangan
- Qashqadaryo, Surxondaryo, Sirdaryo, Jizzax, Navoiy, Xorazm
- Qoraqalpog'iston

### Discounts (Auto-seeded)
- 1 person: 0%
- 2 persons: 10%
- 3 persons: 15%
- 4 persons (full car): 20%

### Users
- SuperAdmin: +998901234567 / admin123

## Production Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for Ubuntu server deployment.

## Full Documentation

- [README.md](README.md) - Complete project documentation
- [API_DOCUMENTATION.md](API_DOCUMENTATION.md) - Detailed API reference
- [DEPLOYMENT.md](DEPLOYMENT.md) - Ubuntu deployment guide

## Need Help?

1. Check logs for errors
2. Review [README.md](README.md)
3. Open an issue on GitHub
4. Contact support

---

**You're all set!** Start building your taxi service. 🚖
