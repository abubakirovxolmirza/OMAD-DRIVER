# Frontend Integration Guide - Taxi Service API

## 📱 Overview

This document provides comprehensive guidance for frontend developers integrating with the Taxi Service API. The API uses JWT-based authentication and follows REST principles.

## 🔌 Base URLs

- **Production**: `https://api.omad-driver.uz/api/v1`
- **Development**: `http://localhost:8080/api/v1`

## 🔐 Authentication

### JWT Token Format

All authenticated requests require a Bearer token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

### Token Expiration

- Default expiration: 720 hours (30 days)
- When expired, users must log in again

### Handling Token Errors

```javascript
// Example error response
{
  "error": "Invalid or expired token"
}

// On 401 Unauthorized, redirect user to login
```

---

## 📚 API Endpoints

### Authentication

#### Register New User
```
POST /auth/register
Content-Type: application/json

{
  "phone_number": "+998901234567",
  "name": "John Doe",
  "password": "secure_password",
  "confirm_password": "secure_password"
}

Response (201):
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "role": "user",
  "user": {
    "id": 1,
    "phone_number": "+998901234567",
    "name": "John Doe",
    "role": "user",
    "language": "uz_latin",
    "avatar": null,
    "is_blocked": false,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Login
```
POST /auth/login
Content-Type: application/json

{
  "phone_number": "+998901234567",
  "password": "secure_password"
}

Response (200):
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "role": "user",
  "user": {
    "id": 1,
    "phone_number": "+998901234567",
    "name": "John Doe",
    "role": "user",
    "language": "uz_latin",
    "avatar": null,
    "is_blocked": false,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}

Response (401):
{
  "error": "Invalid credentials"
}
```

#### Get Profile
```
GET /auth/profile
Authorization: Bearer <token>

Response (200):
{
  "id": 1,
  "phone_number": "+998901234567",
  "name": "John Doe",
  "role": "user",
  "language": "uz_latin",
  "avatar": "/uploads/avatars/uuid_timestamp.jpg",
  "is_blocked": false,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### Update Profile
```
PUT /auth/profile
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Jane Doe",
  "language": "uz_cyrillic"
}

Response (200): (Updated user object)
```

#### Change Password
```
POST /auth/change-password
Authorization: Bearer <token>
Content-Type: application/json

{
  "old_password": "current_password",
  "new_password": "new_password",
  "confirm_new_password": "new_password"
}

Response (200):
{
  "message": "Password changed successfully"
}
```

#### Upload Avatar
```
POST /auth/avatar
Authorization: Bearer <token>
Content-Type: multipart/form-data

Form Data:
  avatar: <file> (JPG, PNG, GIF - max 10MB)

Response (200):
{
  "message": "Avatar uploaded successfully",
  "avatar": "/uploads/avatars/uuid_timestamp.jpg"
}
```

---

### Regions & Districts

#### Get All Regions
```
GET /regions

Response (200):
[
  {
    "id": 1,
    "name_uz_lat": "Toshkent",
    "name_uz_cyr": "Тошкент",
    "name_ru": "Ташкент",
    "created_at": "2024-01-15T10:30:00Z"
  },
  ...
]
```

#### Get Region Details
```
GET /regions/:id

Response (200):
{
  "id": 1,
  "name_uz_lat": "Toshkent",
  "name_uz_cyr": "Тошкент",
  "name_ru": "Ташкент",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Get Districts for Region
```
GET /regions/:id/districts

Response (200):
[
  {
    "id": 1,
    "region_id": 1,
    "name_uz_lat": "Bektemir",
    "name_uz_cyr": "Бектемир",
    "name_ru": "Бектемир",
    "created_at": "2024-01-15T10:30:00Z"
  },
  ...
]
```

#### Get District Details
```
GET /districts/:id

Response (200):
{
  "id": 1,
  "region_id": 1,
  "name_uz_lat": "Bektemir",
  "name_uz_cyr": "Бектемир",
  "name_ru": "Бектемир",
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

### Orders

#### Create Taxi Order
```
POST /orders/taxi
Authorization: Bearer <token>
Content-Type: application/json

{
  "customer_name": "John Doe",
  "customer_phone": "+998901234567",
  "from_region_id": 1,
  "from_district_id": 1,
  "to_region_id": 2,
  "to_district_id": 5,
  "passenger_count": 2,
  "scheduled_date": "2024-01-20",
  "time_range_start": "09:00",
  "time_range_end": "10:00",
  "notes": "Please arrive on time"
}

Response (201):
{
  "id": 1,
  "user_id": 1,
  "driver_id": null,
  "order_type": "taxi",
  "status": "pending",
  "customer_name": "John Doe",
  "customer_phone": "+998901234567",
  "from_region_id": 1,
  "to_region_id": 2,
  "passenger_count": 2,
  "scheduled_date": "2024-01-20",
  "time_range_start": "09:00",
  "time_range_end": "10:00",
  "price": 150000,
  "service_fee": 22500,
  "discount_percentage": 10,
  "final_price": 157500,
  "notes": "Please arrive on time",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### Create Delivery Order
```
POST /orders/delivery
Authorization: Bearer <token>
Content-Type: application/json

{
  "customer_name": "Business Name",
  "customer_phone": "+998901234567",
  "recipient_phone": "+998909876543",
  "from_region_id": 1,
  "from_district_id": 1,
  "to_region_id": 2,
  "to_district_id": 5,
  "delivery_type": "document",
  "scheduled_date": "2024-01-20",
  "time_range_start": "09:00",
  "time_range_end": "17:00",
  "notes": "Fragile items, handle with care"
}

Response (201): (Order object)
```

#### Get My Orders
```
GET /orders/my?status=pending&limit=20&offset=0
Authorization: Bearer <token>

Query Parameters:
  status: pending|accepted|completed|cancelled (optional)
  limit: number of results (default: 20)
  offset: pagination offset (default: 0)

Response (200):
[
  {
    "id": 1,
    "user_id": 1,
    ...
  },
  ...
]
```

#### Get Order Details
```
GET /orders/:id
Authorization: Bearer <token>

Response (200): (Order object)
```

#### Cancel Order
```
POST /orders/:id/cancel
Authorization: Bearer <token>
Content-Type: application/json

{
  "cancellation_reason": "Found alternative transport"
}

Response (200):
{
  "message": "Order cancelled successfully"
}
```

---

### Ratings

#### Submit Rating
```
POST /ratings
Authorization: Bearer <token>
Content-Type: application/json

{
  "order_id": 1,
  "driver_id": 5,
  "rating": 5,
  "comment": "Excellent service, very professional driver"
}

Response (201):
{
  "id": 1,
  "order_id": 1,
  "user_id": 1,
  "driver_id": 5,
  "rating": 5,
  "comment": "Excellent service, very professional driver",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Get Driver Ratings
```
GET /ratings/driver/:driver_id
Authorization: Bearer <token>

Response (200):
[
  {
    "id": 1,
    "order_id": 1,
    "user_id": 1,
    "driver_id": 5,
    "rating": 5,
    "comment": "Excellent service",
    "created_at": "2024-01-15T10:30:00Z"
  },
  ...
]
```

---

### Notifications

#### Get My Notifications
```
GET /notifications?limit=20&offset=0
Authorization: Bearer <token>

Response (200):
[
  {
    "id": 1,
    "user_id": 1,
    "title": "Order Update",
    "message": "Your order has been accepted by a driver",
    "type": "order_status",
    "related_id": 1,
    "is_read": false,
    "created_at": "2024-01-15T10:30:00Z"
  },
  ...
]
```

#### Mark Notification as Read
```
POST /notifications/:id/read
Authorization: Bearer <token>

Response (200):
{
  "message": "Notification marked as read"
}
```

---

### Driver Management

#### Apply as Driver
```
POST /driver/apply
Authorization: Bearer <token>
Content-Type: multipart/form-data

Form Data:
  full_name: "John Doe" (required)
  car_model: "Toyota Camry" (required)
  car_number: "01A001AA" (required)
  license_image: <file> (required, image file)

Response (201):
{
  "id": 1,
  "user_id": 1,
  "full_name": "John Doe",
  "car_model": "Toyota Camry",
  "car_number": "01A001AA",
  "license_image": "/uploads/licenses/uuid_timestamp.jpg",
  "status": "pending",
  "rejection_reason": null,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### Get Driver Profile (Driver Only)
```
GET /driver/profile
Authorization: Bearer <token>

Response (200):
{
  "id": 1,
  "user_id": 1,
  "full_name": "John Doe",
  "car_model": "Toyota Camry",
  "car_number": "01A001AA",
  "license_image": "/uploads/licenses/uuid_timestamp.jpg",
  "balance": 250000,
  "rating": 4.8,
  "total_ratings": 45,
  "status": "approved",
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### Update Driver Profile (Driver Only)
```
PUT /driver/profile
Authorization: Bearer <token>
Content-Type: application/json

{
  "full_name": "John Doe",
  "car_model": "Toyota Camry 2024",
  "car_number": "01A001AA"
}

Response (200): (Updated driver object)
```

#### Get Available Orders (Driver Only)
```
GET /driver/orders/new?limit=20&offset=0
Authorization: Bearer <token>

Response (200):
[
  {
    "id": 1,
    "user_id": 1,
    "order_type": "taxi",
    "status": "pending",
    "customer_name": "John Doe",
    "customer_phone": "+998901234567",
    "from_region_id": 1,
    "to_region_id": 2,
    "passenger_count": 2,
    "price": 150000,
    "final_price": 157500,
    "created_at": "2024-01-15T10:30:00Z"
  },
  ...
]
```

#### Accept Order (Driver Only)
```
POST /driver/orders/:id/accept
Authorization: Bearer <token>

Response (200):
{
  "message": "Order accepted successfully",
  "order_id": 1
}
```

#### Complete Order (Driver Only)
```
POST /driver/orders/:id/complete
Authorization: Bearer <token>
Content-Type: application/json

{
  "notes": "Completed successfully"
}

Response (200):
{
  "message": "Order completed successfully"
}
```

#### Get Driver Orders (Driver Only)
```
GET /driver/orders?status=completed&limit=20&offset=0
Authorization: Bearer <token>

Response (200): (Array of orders)
```

#### Get Driver Statistics (Driver Only)
```
GET /driver/statistics
Authorization: Bearer <token>

Response (200):
{
  "total_completed_orders": 150,
  "total_earnings": 3750000,
  "today_earnings": 250000,
  "this_month_earnings": 1500000,
  "average_rating": 4.8,
  "total_ratings": 45,
  "active_orders": 2
}
```

---

### Feedback

#### Submit Feedback
```
POST /feedback
Authorization: Bearer <token>
Content-Type: application/json

{
  "message": "Great service! Please add more routes to my area."
}

Response (201):
{
  "id": 1,
  "user_id": 1,
  "message": "Great service! Please add more routes to my area.",
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

### Admin Endpoints (Admin/SuperAdmin Only)

#### Get Driver Applications
```
GET /admin/driver-applications?status=pending
Authorization: Bearer <token>

Response (200):
[
  {
    "id": 1,
    "user_id": 1,
    "full_name": "John Doe",
    "phone_number": "+998901234567",
    "car_model": "Toyota Camry",
    "car_number": "01A001AA",
    "license_image": "/uploads/licenses/uuid_timestamp.jpg",
    "status": "pending",
    "rejection_reason": null,
    "reviewed_by": null,
    "reviewed_at": null,
    "created_at": "2024-01-15T10:30:00Z"
  },
  ...
]
```

#### Review Driver Application
```
POST /admin/driver-applications/:id/review
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "approved",
  "rejection_reason": ""
}

OR

{
  "status": "rejected",
  "rejection_reason": "License image is not clear"
}

Response (200):
{
  "message": "Application reviewed successfully"
}
```

#### Get All Drivers
```
GET /admin/drivers?status=approved
Authorization: Bearer <token>

Response (200): (Array of driver objects)
```

#### Add Driver Balance
```
POST /admin/drivers/:id/add-balance
Authorization: Bearer <token>
Content-Type: application/json

{
  "amount": 500000
}

Response (200):
{
  "message": "Balance added successfully"
}
```

#### Block/Unblock User
```
POST /admin/users/:id/block
Authorization: Bearer <token>
Content-Type: application/json

{
  "is_blocked": true
}

Response (200):
{
  "message": "User blocked successfully"
}
```

#### Get All Orders (Admin)
```
GET /admin/orders?status=completed&type=taxi&from_date=2024-01-01&to_date=2024-01-31
Authorization: Bearer <token>

Query Parameters:
  status: pending|accepted|completed|cancelled
  type: taxi|delivery
  from_date: YYYY-MM-DD
  to_date: YYYY-MM-DD

Response (200): (Array of orders)
```

#### Get Platform Statistics
```
GET /admin/statistics
Authorization: Bearer <token>

Response (200):
{
  "total_users": 1250,
  "total_drivers": 180,
  "active_drivers": 145,
  "total_orders": 5680,
  "completed_orders": 5420,
  "total_revenue": 142050000,
  "today_orders": 42,
  "today_revenue": 1050000
}
```

#### Get All Pricing
```
GET /admin/pricing
Authorization: Bearer <token>

Response (200):
[
  {
    "id": 1,
    "from_region_id": 1,
    "to_region_id": 2,
    "base_price": 150000,
    "price_per_person": 10000,
    "service_fee": 15,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  ...
]
```

#### Set Pricing Route
```
POST /admin/pricing
Authorization: Bearer <token>
Content-Type: application/json

{
  "from_region_id": 1,
  "to_region_id": 2,
  "base_price": 150000,
  "price_per_person": 10000,
  "service_fee": 15
}

Response (200): (Pricing object)
```

#### Get Feedback (Admin)
```
GET /admin/feedback
Authorization: Bearer <token>

Response (200):
[
  {
    "id": 1,
    "user_id": 1,
    "message": "Great service!",
    "created_at": "2024-01-15T10:30:00Z"
  },
  ...
]
```

---

### SuperAdmin Endpoints

#### Create Admin User
```
POST /admin/create-admin
Authorization: Bearer <superadmin_token>
Content-Type: application/json

{
  "phone_number": "+998901111111",
  "name": "Admin User",
  "password": "secure_password"
}

Response (201): (User object)
```

#### Reset User Password
```
POST /admin/users/:id/reset-password
Authorization: Bearer <superadmin_token>
Content-Type: application/json

{
  "new_password": "new_secure_password"
}

Response (200):
{
  "message": "Password reset successfully"
}
```

---

## 🛠️ Frontend Implementation Tips

### 1. Session Management

```javascript
// Store token securely
localStorage.setItem('authToken', response.token);
localStorage.setItem('userRole', response.role);

// Clear on logout
localStorage.removeItem('authToken');
localStorage.removeItem('userRole');
```

### 2. API Request Helper

```javascript
async function apiRequest(endpoint, method = 'GET', data = null) {
  const token = localStorage.getItem('authToken');
  const headers = {
    'Content-Type': 'application/json',
  };
  
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const options = {
    method,
    headers,
  };

  if (data) {
    options.body = JSON.stringify(data);
  }

  const response = await fetch(`https://api.omad-driver.uz/api/v1${endpoint}`, options);

  if (response.status === 401) {
    // Token expired or invalid
    localStorage.removeItem('authToken');
    // Redirect to login
    window.location.href = '/login';
  }

  return await response.json();
}
```

### 3. Role-Based Access Control

```javascript
const userRole = localStorage.getItem('userRole');

// Show/hide UI elements based on role
if (userRole === 'driver') {
  showDriverDashboard();
} else if (userRole === 'admin') {
  showAdminDashboard();
} else {
  showUserDashboard();
}
```

### 4. Error Handling

```javascript
async function handleApiError(error) {
  if (error.error) {
    if (error.error.includes('Insufficient permissions')) {
      showError('You do not have permission to access this resource');
    } else if (error.error.includes('Invalid credentials')) {
      showError('Phone number or password is incorrect');
    } else {
      showError(error.error);
    }
  }
}
```

### 5. File Upload

```javascript
async function uploadAvatar(file) {
  const formData = new FormData();
  formData.append('avatar', file);

  const token = localStorage.getItem('authToken');
  const response = await fetch('https://api.omad-driver.uz/api/v1/auth/avatar', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
    body: formData,
  });

  return await response.json();
}
```

---

## 📋 Language Support

The API supports three languages for region and district names:

- `uz_latin`: Uzbek (Latin script)
- `uz_cyrillic`: Uzbek (Cyrillic script)
- `ru`: Russian

Users can select their preferred language in their profile.

---

## ⏱️ Rate Limiting

Currently, the API does not have rate limiting, but it's recommended to implement client-side request throttling to avoid abuse.

---

## 🔄 WebSocket Support (Future)

Real-time order tracking and driver location updates will be available via WebSocket connection in future versions.

---

## 📞 Support

For API issues or questions, contact: support@taxiservice.com

---

## 🚀 Deployment Checklist

Before deploying to production:

- [ ] Update API base URL in frontend
- [ ] Enable HTTPS/SSL
- [ ] Test all authentication flows
- [ ] Verify role-based access control
- [ ] Test file uploads
- [ ] Load test the API
- [ ] Set up error logging/monitoring
- [ ] Configure CORS properly for your domain

---

## Changelog

### Version 1.0 (2024-01-15)
- Initial API release
- User authentication and profile management
- Taxi and delivery orders
- Driver management and ratings
- Admin dashboard
- Region and district management
