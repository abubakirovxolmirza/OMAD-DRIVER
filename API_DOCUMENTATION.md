# API Documentation

Complete API reference for Taxi Service Backend.

**Base URL**: `http://localhost:8080/api/v1`

**Swagger UI**: `http://localhost:8080/swagger/index.html`

## Authentication

All protected endpoints require a JWT token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

## Response Format

### Success Response
```json
{
  "data": { ... },
  "message": "Success message"
}
```

### Error Response
```json
{
  "error": "Error message description"
}
```

---

## Authentication Endpoints

### Register User

Create a new user account.

**Endpoint**: `POST /auth/register`

**Request Body**:
```json
{
  "phone_number": "+998901234567",
  "name": "John Doe",
  "password": "securePassword123",
  "confirm_password": "securePassword123"
}
```

**Response** (201 Created):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "phone_number": "+998901234567",
    "name": "John Doe",
    "role": "user",
    "language": "uz_latin",
    "avatar": null,
    "is_blocked": false,
    "created_at": "2025-11-03T10:00:00Z",
    "updated_at": "2025-11-03T10:00:00Z"
  }
}
```

**Errors**:
- `400` - Validation error or passwords don't match
- `409` - Phone number already registered

---

### Login

Authenticate and receive JWT token.

**Endpoint**: `POST /auth/login`

**Request Body**:
```json
{
  "phone_number": "+998901234567",
  "password": "securePassword123"
}
```

**Response** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "phone_number": "+998901234567",
    "name": "John Doe",
    "role": "user",
    "language": "uz_latin",
    "avatar": null,
    "is_blocked": false,
    "created_at": "2025-11-03T10:00:00Z",
    "updated_at": "2025-11-03T10:00:00Z"
  }
}
```

**Errors**:
- `400` - Invalid request
- `401` - Invalid credentials
- `403` - Account is blocked

---

### Get Profile

Get current user's profile.

**Endpoint**: `GET /auth/profile`

**Headers**: `Authorization: Bearer <token>`

**Response** (200 OK):
```json
{
  "id": 1,
  "phone_number": "+998901234567",
  "name": "John Doe",
  "role": "user",
  "language": "uz_latin",
  "avatar": "avatars/uuid_timestamp.jpg",
  "is_blocked": false,
  "created_at": "2025-11-03T10:00:00Z",
  "updated_at": "2025-11-03T10:00:00Z"
}
```

---

### Update Profile

Update user's name and language.

**Endpoint**: `PUT /auth/profile`

**Headers**: `Authorization: Bearer <token>`

**Request Body**:
```json
{
  "name": "John Updated",
  "language": "ru"
}
```

**Languages**: `uz_latin`, `uz_cyrillic`, `ru`

**Response** (200 OK): Updated user object

---

### Change Password

Change user's password.

**Endpoint**: `POST /auth/change-password`

**Headers**: `Authorization: Bearer <token>`

**Request Body**:
```json
{
  "old_password": "oldPassword123",
  "new_password": "newPassword456",
  "confirm_new_password": "newPassword456"
}
```

**Response** (200 OK):
```json
{
  "message": "Password changed successfully"
}
```

**Errors**:
- `400` - Passwords don't match or invalid old password

---

### Upload Avatar

Upload profile avatar image.

**Endpoint**: `POST /auth/avatar`

**Headers**: `Authorization: Bearer <token>`

**Content-Type**: `multipart/form-data`

**Form Data**:
- `avatar`: Image file (jpg, jpeg, png, gif - max 10MB)

**Response** (200 OK):
```json
{
  "message": "Avatar uploaded successfully",
  "avatar": "avatars/uuid_timestamp.jpg"
}
```

**Errors**:
- `400` - No file or file too large
- `400` - Invalid file type

---

## Order Endpoints

### Create Taxi Order

Create a new taxi order.

**Endpoint**: `POST /orders/taxi`

**Headers**: `Authorization: Bearer <token>`

**Request Body**:
```json
{
  "customer_name": "John Doe",
  "customer_phone": "+998901234567",
  "from_region_id": 1,
  "from_district_id": 5,
  "to_region_id": 2,
  "to_district_id": 12,
  "passenger_count": 2,
  "scheduled_date": "15.11.2025",
  "time_range_start": "09:00",
  "time_range_end": "11:00",
  "notes": "Please call before arrival"
}
```

**Passenger Count**: 1, 2, 3, or 4 (4 = full car)

**Response** (201 Created):
```json
{
  "id": 1,
  "user_id": 1,
  "driver_id": null,
  "order_type": "taxi",
  "status": "pending",
  "customer_name": "John Doe",
  "customer_phone": "+998901234567",
  "from_region_id": 1,
  "from_district_id": 5,
  "to_region_id": 2,
  "to_district_id": 12,
  "passenger_count": 2,
  "scheduled_date": "2025-11-15T00:00:00Z",
  "time_range_start": "09:00",
  "time_range_end": "11:00",
  "price": 150000,
  "service_fee": 22500,
  "discount_percentage": 10,
  "final_price": 157500,
  "notes": "Please call before arrival",
  "created_at": "2025-11-03T10:00:00Z",
  "updated_at": "2025-11-03T10:00:00Z"
}
```

**Pricing Logic**:
1. Base price + (price per person × passenger count)
2. Apply discount based on passenger count
3. Add service fee (percentage)

**Discounts**:
- 1 person: 0%
- 2 persons: 10%
- 3 persons: 15%
- 4 persons (full car): 20%

**Errors**:
- `400` - Invalid date format or validation error
- `400` - From and To regions must be different
- `400` - Pricing not configured for route

---

### Create Delivery Order

Create a new delivery order.

**Endpoint**: `POST /orders/delivery`

**Headers**: `Authorization: Bearer <token>`

**Request Body**:
```json
{
  "customer_name": "John Doe",
  "customer_phone": "+998901234567",
  "recipient_phone": "+998907654321",
  "from_region_id": 1,
  "from_district_id": 5,
  "to_region_id": 2,
  "to_district_id": 12,
  "delivery_type": "document",
  "scheduled_date": "15.11.2025",
  "time_range_start": "09:00",
  "time_range_end": "11:00",
  "notes": "Fragile items"
}
```

**Delivery Types**: `document`, `box`, `luggage`, `valuable`, `other`

**Response** (201 Created): Similar to taxi order

---

### Get My Orders

Get all orders created by current user.

**Endpoint**: `GET /orders/my`

**Headers**: `Authorization: Bearer <token>`

**Query Parameters**:
- `status` (optional): Filter by status - `pending`, `accepted`, `in_progress`, `completed`, `cancelled`
- `type` (optional): Filter by type - `taxi`, `delivery`

**Example**: `/orders/my?status=completed&type=taxi`

**Response** (200 OK): Array of order objects

---

### Get Order Details

Get detailed information about a specific order.

**Endpoint**: `GET /orders/:id`

**Headers**: `Authorization: Bearer <token>`

**Example**: `/orders/123`

**Response** (200 OK): Order object with all details

**Note**: Users can only see their own orders. Drivers and admins can see any order.

---

### Cancel Order

Cancel a pending or accepted order.

**Endpoint**: `POST /orders/:id/cancel`

**Headers**: `Authorization: Bearer <token>`

**Request Body**:
```json
{
  "reason": "Changed my plans"
}
```

**Response** (200 OK):
```json
{
  "message": "Order cancelled successfully"
}
```

**Behavior**:
- If order was accepted by driver, service fee is refunded
- Cancellation notification sent to Telegram admin group
- Driver is notified

**Errors**:
- `400` - Cannot cancel order in current status
- `404` - Order not found

---

## Driver Endpoints

### Apply as Driver

Submit application to become a driver.

**Endpoint**: `POST /driver/apply`

**Headers**: `Authorization: Bearer <token>`

**Content-Type**: `multipart/form-data`

**Form Data**:
- `full_name`: Full name
- `car_model`: Car model (e.g., "Chevrolet Lacetti")
- `car_number`: Car number (e.g., "01A123BC")
- `license_image`: Driver's license image file

**Response** (201 Created):
```json
{
  "id": 1,
  "user_id": 1,
  "full_name": "John Driver",
  "phone_number": "+998901234567",
  "car_model": "Chevrolet Lacetti",
  "car_number": "01A123BC",
  "license_image": "licenses/uuid_timestamp.jpg",
  "status": "pending",
  "created_at": "2025-11-03T10:00:00Z",
  "updated_at": "2025-11-03T10:00:00Z"
}
```

**Errors**:
- `400` - Already a driver or missing required fields
- `409` - Application already pending

---

### Get Driver Profile

Get driver's profile information.

**Endpoint**: `GET /driver/profile`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Driver

**Response** (200 OK):
```json
{
  "id": 1,
  "user_id": 1,
  "full_name": "John Driver",
  "car_model": "Chevrolet Lacetti",
  "car_number": "01A123BC",
  "license_image": "licenses/uuid_timestamp.jpg",
  "balance": 150000,
  "rating": 4.8,
  "total_ratings": 24,
  "status": "approved",
  "is_active": true,
  "created_at": "2025-11-03T10:00:00Z",
  "updated_at": "2025-11-03T10:00:00Z"
}
```

---

### Update Driver Profile

Update driver's profile.

**Endpoint**: `PUT /driver/profile`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Driver

**Request Body**:
```json
{
  "full_name": "John Updated Driver",
  "car_model": "Chevrolet Nexia",
  "car_number": "01B456CD"
}
```

**Response** (200 OK): Updated driver object

---

### Get New Orders

Get available orders for drivers to accept.

**Endpoint**: `GET /driver/orders/new`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Driver

**Query Parameters**:
- `type` (optional): `taxi` or `delivery`
- `from_region` (optional): Filter by from region ID
- `to_region` (optional): Filter by to region ID

**Example**: `/driver/orders/new?type=taxi&from_region=1`

**Response** (200 OK): Array of pending orders

---

### Accept Order

Accept an available order.

**Endpoint**: `POST /driver/orders/:id/accept`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Driver

**Response** (200 OK): Accepted order object

**Behavior**:
- Service fee is deducted from driver's balance
- Order status changes to `accepted`
- User is notified
- Driver has 5 minutes to accept from order creation

**Errors**:
- `400` - Insufficient balance
- `400` - Order no longer available or deadline passed
- `403` - Driver account not active

---

### Complete Order

Mark an order as completed.

**Endpoint**: `POST /driver/orders/:id/complete`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Driver

**Response** (200 OK):
```json
{
  "message": "Order completed successfully"
}
```

**Behavior**:
- Order status changes to `completed`
- User is notified to rate the driver

**Errors**:
- `400` - Order not assigned to you or not in accepted status

---

### Get Driver Orders

Get all orders assigned to the driver.

**Endpoint**: `GET /driver/orders`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Driver

**Query Parameters**:
- `status` (optional): Filter by status

**Response** (200 OK): Array of order objects

---

### Get Driver Statistics

Get driver's performance statistics.

**Endpoint**: `GET /driver/statistics`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Driver

**Query Parameters**:
- `period` (optional): `daily`, `monthly`, `yearly` (default: all time)

**Response** (200 OK):
```json
{
  "total_orders": 45,
  "completed_orders": 42,
  "total_earnings": 630000,
  "current_balance": 85000,
  "average_rating": 4.8,
  "total_ratings": 24
}
```

---

## Admin Endpoints

All admin endpoints require Admin or SuperAdmin role.

### Get Driver Applications

Get list of driver applications.

**Endpoint**: `GET /admin/driver-applications`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Admin, SuperAdmin

**Query Parameters**:
- `status` (optional): `pending`, `approved`, `rejected`

**Response** (200 OK): Array of application objects

---

### Review Driver Application

Approve or reject a driver application.

**Endpoint**: `POST /admin/driver-applications/:id/review`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Admin, SuperAdmin

**Request Body**:
```json
{
  "status": "approved",
  "rejection_reason": "Optional rejection reason"
}
```

**Status**: `approved` or `rejected`

**Response** (200 OK):
```json
{
  "message": "Application reviewed successfully"
}
```

**Behavior**:
- If approved: User role changes to driver, driver profile created
- If rejected: Application marked as rejected with reason
- User is notified of the decision

---

### Get All Drivers

Get list of all drivers.

**Endpoint**: `GET /admin/drivers`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Admin, SuperAdmin

**Query Parameters**:
- `status` (optional): Filter by status

**Response** (200 OK): Array of driver objects

---

### Add Driver Balance

Add balance to a driver's account.

**Endpoint**: `POST /admin/drivers/:id/add-balance`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Admin, SuperAdmin

**Request Body**:
```json
{
  "amount": 50000
}
```

**Response** (200 OK):
```json
{
  "message": "Balance added successfully"
}
```

**Behavior**:
- Balance is credited to driver
- Transaction record is created

---

### Block/Unblock User

Block or unblock a user or driver.

**Endpoint**: `POST /admin/users/:id/block`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Admin, SuperAdmin

**Request Body**:
```json
{
  "is_blocked": true
}
```

**Response** (200 OK):
```json
{
  "message": "User blocked successfully"
}
```

---

### Set Pricing

Set or update pricing for a route.

**Endpoint**: `POST /admin/pricing`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Admin, SuperAdmin

**Request Body**:
```json
{
  "from_region_id": 1,
  "to_region_id": 2,
  "base_price": 100000,
  "price_per_person": 25000,
  "service_fee": 15
}
```

**Response** (200 OK): Pricing object

**Note**: If pricing exists for route, it will be updated.

---

### Get All Pricing

Get all configured pricing routes.

**Endpoint**: `GET /admin/pricing`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Admin, SuperAdmin

**Response** (200 OK): Array of pricing objects

---

### Get All Orders

Get all orders with advanced filters.

**Endpoint**: `GET /admin/orders`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Admin, SuperAdmin

**Query Parameters**:
- `status` (optional): Filter by status
- `type` (optional): Filter by type
- `from_date` (optional): From date (YYYY-MM-DD)
- `to_date` (optional): To date (YYYY-MM-DD)

**Example**: `/admin/orders?status=completed&from_date=2025-11-01&to_date=2025-11-30`

**Response** (200 OK): Array of order objects

---

### Get Platform Statistics

Get overall platform statistics.

**Endpoint**: `GET /admin/statistics`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Admin, SuperAdmin

**Response** (200 OK):
```json
{
  "total_users": 1250,
  "total_drivers": 85,
  "active_drivers": 72,
  "total_orders": 3420,
  "completed_orders": 3180,
  "total_revenue": 47700000,
  "today_orders": 45,
  "today_revenue": 675000
}
```

---

### Get All Feedback

Get all user feedback/suggestions.

**Endpoint**: `GET /admin/feedback`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: Admin, SuperAdmin

**Response** (200 OK): Array of feedback objects

---

## SuperAdmin Endpoints

### Create Admin

Create a new admin user.

**Endpoint**: `POST /admin/create-admin`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: SuperAdmin only

**Request Body**:
```json
{
  "phone_number": "+998901234567",
  "name": "New Admin",
  "password": "adminPassword123"
}
```

**Response** (201 Created): Created admin user object

**Errors**:
- `409` - Phone number already registered

---

### Reset User Password

Reset a user's password.

**Endpoint**: `POST /admin/users/:id/reset-password`

**Headers**: `Authorization: Bearer <token>`

**Role Required**: SuperAdmin only

**Request Body**:
```json
{
  "new_password": "newPassword123"
}
```

**Response** (200 OK):
```json
{
  "message": "Password reset successfully"
}
```

---

## Rating Endpoints

### Create Rating

Rate a driver after order completion.

**Endpoint**: `POST /ratings`

**Headers**: `Authorization: Bearer <token>`

**Request Body**:
```json
{
  "order_id": 123,
  "rating": 5,
  "comment": "Great driver, very professional!"
}
```

**Rating**: 1-5 stars

**Response** (201 Created): Rating object

**Errors**:
- `400` - Can only rate completed orders
- `409` - Order already rated
- `404` - Order not found

---

### Get Driver Ratings

Get all ratings for a specific driver.

**Endpoint**: `GET /ratings/driver/:driver_id`

**Response** (200 OK): Array of rating objects

---

## Notification Endpoints

### Get My Notifications

Get all notifications for current user.

**Endpoint**: `GET /notifications`

**Headers**: `Authorization: Bearer <token>`

**Query Parameters**:
- `unread` (optional): Set to `true` to get only unread notifications

**Response** (200 OK): Array of notification objects

---

### Mark Notification as Read

Mark a notification as read.

**Endpoint**: `POST /notifications/:id/read`

**Headers**: `Authorization: Bearer <token>`

**Response** (200 OK):
```json
{
  "message": "Notification marked as read"
}
```

---

## Region Endpoints

### Get All Regions

Get list of all regions (provinces).

**Endpoint**: `GET /regions`

**No authentication required**

**Response** (200 OK):
```json
[
  {
    "id": 1,
    "name_uz_lat": "Toshkent",
    "name_uz_cyr": "Тошкент",
    "name_ru": "Ташкент",
    "created_at": "2025-11-03T10:00:00Z"
  }
]
```

---

### Get Districts by Region

Get all districts for a specific region.

**Endpoint**: `GET /regions/:region_id/districts`

**No authentication required**

**Response** (200 OK): Array of district objects

---

## Feedback Endpoints

### Submit Feedback

Submit feedback or suggestion.

**Endpoint**: `POST /feedback`

**Headers**: `Authorization: Bearer <token>`

**Request Body**:
```json
{
  "message": "Great service! Would be nice to have a mobile app."
}
```

**Response** (201 Created): Feedback object

---

## HTTP Status Codes

- `200` - OK
- `201` - Created
- `400` - Bad Request (validation error)
- `401` - Unauthorized (invalid or missing token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `409` - Conflict (resource already exists)
- `500` - Internal Server Error

## Rate Limiting

Currently no rate limiting is implemented. Consider adding rate limiting in production.

## Pagination

Currently pagination is not implemented. All list endpoints return all results. Consider adding pagination for large datasets.

## Testing with cURL

### Register
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"phone_number":"+998901234567","name":"Test User","password":"test123","confirm_password":"test123"}'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"phone_number":"+998901234567","password":"test123"}'
```

### Get Profile
```bash
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Swagger Documentation

For interactive API documentation and testing, visit:

```
http://localhost:8080/swagger/index.html
```

The Swagger UI provides:
- Complete API reference
- Request/response examples
- Interactive API testing
- Model schemas
- Authentication testing
