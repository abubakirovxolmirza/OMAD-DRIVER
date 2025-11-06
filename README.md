# Taxi Service Backend API

A professional taxi and delivery service backend built with Go, PostgreSQL, and REST API architecture.

## Features

### User Features
- **Registration & Authentication** - JWT-based authentication with phone number
- **Profile Management** - Update profile, change password, upload avatar
- **Multi-language Support** - Uzbek (Latin & Cyrillic), Russian
- **Taxi Orders** - Create taxi orders with automatic pricing and discounts
- **Delivery Orders** - Send packages/documents between regions
- **Order Management** - View order history, track active orders, cancel orders
- **Driver Rating** - Rate drivers after completed trips

### Driver Features
- **Driver Application** - Apply to become a driver with license verification
- **Order Management** - View new orders, accept orders, complete trips
- **Balance System** - Service fee deduction, balance tracking
- **Statistics** - Daily, monthly, yearly earnings and order statistics
- **Rating System** - Receive ratings from customers

### Admin Features
- **Driver Approval** - Review and approve/reject driver applications
- **User Management** - Block/unblock users and drivers
- **Pricing Configuration** - Set prices between regions
- **Order Reports** - View all orders with filters
- **Statistics Dashboard** - Platform-wide statistics
- **Balance Management** - Add balance to driver accounts

### SuperAdmin Features
- All admin features plus:
- **Admin Management** - Create new admin users
- **Password Reset** - Reset user passwords

## Technology Stack

- **Language**: Go 1.21+
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL 12+
- **Authentication**: JWT (JSON Web Tokens)
- **Password Hashing**: bcrypt
- **API Documentation**: Swagger/OpenAPI
- **File Upload**: Multipart form-data

## Project Structure

```
taxi-service/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── database/
│   │   └── database.go         # Database connection and schema
│   ├── handlers/
│   │   ├── auth.go             # Authentication handlers
│   │   ├── order.go            # Order management handlers
│   │   ├── driver.go           # Driver handlers
│   │   ├── admin.go            # Admin handlers
│   │   └── misc.go             # Miscellaneous handlers
│   ├── middleware/
│   │   ├── auth.go             # Authentication middleware
│   │   └── cors.go             # CORS middleware
│   ├── models/
│   │   └── models.go           # Data models
│   └── utils/
│       ├── jwt.go              # JWT utilities
│       ├── password.go         # Password hashing
│       └── file.go             # File upload utilities
├── uploads/                    # Upload directory
├── .env                        # Environment variables (create from .env.example)
├── .env.example               # Example environment variables
├── go.mod                     # Go module definition
├── go.sum                     # Go module checksums
├── Makefile                   # Build and run commands
├── README.md                  # This file
├── API_DOCUMENTATION.md       # Detailed API documentation
└── DEPLOYMENT.md              # Ubuntu deployment guide
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Git

## Installation & Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd TAXI
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment

Copy the example environment file and update with your settings:

```bash
cp .env.example .env
```

Edit `.env` and configure:
- Database credentials
- JWT secret key
- Server port
- Upload directory
- Telegram bot token (optional)

### 4. Setup PostgreSQL Database

Create a new PostgreSQL database:

```bash
createdb taxi_service
```

Or using psql:

```sql
CREATE DATABASE taxi_service;
```

### 5. Run the Application

The application will automatically create all necessary tables on first run:

```bash
go run cmd/main.go
```

Or using Make:

```bash
make run
```

### 6. Access the Application

- **API Base URL**: http://localhost:8080/api/v1
- **Swagger Documentation**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health

### Default SuperAdmin Credentials (Development Only)

- **Phone**: +998901234567
- **Password**: admin123

⚠️ **Important**: Change these credentials immediately in production!

## Development

### Generate Swagger Documentation

```bash
make swagger
```

Or manually:

```bash
swag init -g cmd/main.go -o ./docs
```

### Run Tests

```bash
make test
```

### Build for Production

```bash
make build
```

This creates a `taxi-service` binary in the project root.

## API Endpoints Overview

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login
- `GET /api/v1/auth/profile` - Get profile
- `PUT /api/v1/auth/profile` - Update profile
- `POST /api/v1/auth/change-password` - Change password
- `POST /api/v1/auth/avatar` - Upload avatar

### Orders
- `POST /api/v1/orders/taxi` - Create taxi order
- `POST /api/v1/orders/delivery` - Create delivery order
- `GET /api/v1/orders/my` - Get my orders
- `GET /api/v1/orders/:id` - Get order details
- `POST /api/v1/orders/:id/cancel` - Cancel order

### Driver
- `POST /api/v1/driver/apply` - Apply as driver
- `GET /api/v1/driver/profile` - Get driver profile
- `PUT /api/v1/driver/profile` - Update driver profile
- `GET /api/v1/driver/orders/new` - Get available orders
- `POST /api/v1/driver/orders/:id/accept` - Accept order
- `POST /api/v1/driver/orders/:id/complete` - Complete order
- `GET /api/v1/driver/orders` - Get driver orders
- `GET /api/v1/driver/statistics` - Get statistics

### Admin
- `GET /api/v1/admin/driver-applications` - Get applications
- `POST /api/v1/admin/driver-applications/:id/review` - Review application
- `GET /api/v1/admin/drivers` - Get all drivers
- `POST /api/v1/admin/drivers/:id/add-balance` - Add balance
- `POST /api/v1/admin/users/:id/block` - Block/unblock user
- `POST /api/v1/admin/pricing` - Set pricing
- `GET /api/v1/admin/pricing` - Get pricing
- `GET /api/v1/admin/orders` - Get all orders
- `GET /api/v1/admin/statistics` - Get statistics

See [API_DOCUMENTATION.md](API_DOCUMENTATION.md) for complete API reference.

## Database Schema

### Main Tables
- **users** - User accounts (customers, drivers, admins)
- **drivers** - Driver-specific information
- **orders** - Taxi and delivery orders
- **regions** - Regions/provinces
- **districts** - Districts within regions
- **pricing** - Route pricing configuration
- **discounts** - Passenger count discounts
- **ratings** - Driver ratings
- **notifications** - User notifications
- **driver_applications** - Driver application requests
- **transactions** - Balance transactions
- **feedback** - User feedback/suggestions

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8080` |
| `SERVER_HOST` | HTTP server host | `0.0.0.0` |
| `ENV` | Environment (development/production) | `development` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | `taxi_service` |
| `JWT_SECRET` | JWT signing secret | - |
| `JWT_EXPIRATION_HOURS` | JWT token expiration | `720` (30 days) |
| `UPLOAD_DIR` | File upload directory | `./uploads` |
| `MAX_UPLOAD_SIZE` | Max file size in bytes | `10485760` (10MB) |

## Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed Ubuntu server deployment instructions.

## Security Considerations

1. **Change Default Credentials**: Immediately change the default superadmin password in production
2. **Use Strong JWT Secret**: Generate a strong random JWT secret key
3. **Enable HTTPS**: Use nginx or similar as reverse proxy with SSL/TLS
4. **Database Security**: Use strong database passwords and restrict access
5. **File Upload**: Validate file types and sizes to prevent abuse
6. **Rate Limiting**: Consider adding rate limiting for API endpoints
7. **Backup**: Implement regular database backups

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.

## Support

For support, email support@taxiservice.com or create an issue in the repository.

## Roadmap

- [ ] Telegram Bot Integration
- [ ] Push Notifications (FCM)
- [ ] Real-time Order Tracking (WebSocket)
- [ ] Payment Gateway Integration
- [ ] Mobile App APIs (iOS/Android)
- [ ] Excel Report Generation
- [ ] SMS Notifications
- [ ] Email Notifications
- [ ] Advanced Analytics Dashboard

## Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [PostgreSQL](https://www.postgresql.org/)
- [Swagger](https://swagger.io/)
- All contributors and users of this project
