# Quick Start Guide - Taxi Service API

## 🚀 Get Running in 5 Minutes

### Prerequisites
- Docker & Docker Compose installed (recommended)
- OR Go 1.21+ and PostgreSQL 15+
- Git installed

### Option 1: Docker (Easiest & Recommended)

```bash
# 1. Clone the repository
git clone https://github.com/abubakirovxolmirza/OMAD-DRIVER.git
cd OMAD-DRIVER

# 2. Copy and configure environment
cp .env.example .env
# Edit .env with your settings (change JWT_SECRET, DB_PASSWORD)

# 3. Start services
docker-compose up -d

# 4. Seed database with regions and pricing
docker-compose exec app /app/cmd/tools/dbseed/main.go -action=seed

# 5. Test the API
curl http://localhost:8080/health
```

**Access Points:**
- API: http://localhost:8080/api/v1
- Health Check: http://localhost:8080/health
- Database: localhost:5432

### Option 2: Local Development Setup

```bash
# 1. Clone repository
git clone https://github.com/abubakirovxolmirza/OMAD-DRIVER.git
cd OMAD-DRIVER

# 2. Install dependencies
go mod download

# 3. Create database
createdb taxi_service

# 4. Configure environment
cp .env.example .env
# Edit .env with your database credentials

# 5. Run application
go run cmd/main.go

# 6. In another terminal, seed database
go run cmd/tools/dbseed/main.go -action=seed

# 7. Test the API
curl http://localhost:8080/health
```

---

## 📝 Quick API Tests

### 1. Register a New User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+998901234567",
    "name": "John Doe",
    "password": "SecurePass123",
    "confirm_password": "SecurePass123"
  }'
```

Response includes `token` and `role`:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "role": "user",
  "user": {...}
}
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+998901234567",
    "password": "SecurePass123"
  }'
```

### 3. Get Regions (No Auth Required)

```bash
curl http://localhost:8080/api/v1/regions
```

### 4. Create Taxi Order (Auth Required)

```bash
TOKEN="your_jwt_token_from_login"

curl -X POST http://localhost:8080/api/v1/orders/taxi \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "John Doe",
    "customer_phone": "+998901234567",
    "from_region_id": 1,
    "from_district_id": 1,
    "to_region_id": 2,
    "to_district_id": 5,
    "passenger_count": 2,
    "scheduled_date": "2024-02-15",
    "time_range_start": "09:00",
    "time_range_end": "10:00"
  }'
```

---

## 🎯 Make Commands

```bash
make help               # Show all available commands
make build            # Build the application
make run              # Run the application
make test             # Run tests
make docker-up        # Start Docker containers
make docker-down      # Stop Docker containers
make seed-db          # Populate database
make clean            # Clean build files
```

---

## 🔐 Default Development Credentials

SuperAdmin (created automatically in development):
- **Phone**: +998901234567
- **Password**: admin123

⚠️ **Change immediately in production!**

---

## 📖 Full Documentation

- **Frontend Integration**: `FRONTEND_INTEGRATION_GUIDE.md`
- **Production Deployment**: `PRODUCTION_DEPLOYMENT.md`
- **Complete README**: `README.md`
- **API Documentation**: Generated Swagger at `/swagger/index.html`

---

## 🐳 Docker Quick Reference

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f app

# Check status
docker-compose ps

# Access database
docker-compose exec db psql -U taxi_user -d taxi_service
```

---

## 🔧 Troubleshooting

**API not responding?**
```bash
curl http://localhost:8080/health
docker-compose logs app
```

**Database connection error?**
```bash
docker-compose exec app /bin/sh -c 'nc -zv db 5432'
```

**Port 8080 already in use?**
```bash
# Edit .env to use different port
SERVER_PORT=8081
docker-compose down
docker-compose up -d
```

**Need to reseed database?**
```bash
docker-compose exec app go run cmd/tools/dbseed/main.go -action=cleanup
docker-compose exec app go run cmd/tools/dbseed/main.go -action=seed
```

---

## 🚀 Next Steps

1. **Read Documentation**
   - See `FRONTEND_INTEGRATION_GUIDE.md` for all API endpoints
   - See `PRODUCTION_DEPLOYMENT.md` for production setup

2. **Implement Frontend**
   - Configure API base URL
   - Implement authentication
   - Build UI components

3. **Deploy to Production**
   - Follow production deployment guide
   - Configure domain and SSL
   - Set up monitoring

---

## 💡 Project Structure

```
├── cmd/              # Application entry points
├── internal/         # Core application code
│   ├── config/       # Configuration
│   ├── database/     # DB setup & migrations
│   ├── handlers/     # API handlers
│   ├── middleware/   # Auth & CORS
│   ├── models/       # Data models
│   └── utils/        # Utilities
├── docs/             # API documentation
├── uploads/          # File uploads
├── database/         # Migrations
└── docker-compose.yml
```

---

## 🆘 Need Help?

- Check `README.md` for detailed documentation
- Review `FRONTEND_INTEGRATION_GUIDE.md` for API endpoints
- See `PRODUCTION_DEPLOYMENT.md` for deployment help
- Contact: support@taxiservice.com

---

**Ready to go! 🎉**
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
