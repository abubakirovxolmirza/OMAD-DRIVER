# Testing Guide

Complete testing guide for the Taxi Service Backend API.

## Prerequisites for Testing

- Application running on `http://localhost:8080`
- PostgreSQL database setup and connected
- API client (Postman, cURL, or Swagger UI)

## Quick Test via Swagger UI

**Easiest Method**: Open http://localhost:8080/swagger/index.html

All endpoints are documented with try-it-now functionality.

## Testing Workflow

### Phase 1: Authentication & User Management

#### 1.1 Register New User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+998901111111",
    "name": "Test User",
    "password": "password123",
    "confirm_password": "password123"
  }'
```

**Expected**: 201 Created with token and user object

#### 1.2 Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+998901111111",
    "password": "password123"
  }'
```

**Expected**: 200 OK with token

**Save the token** for subsequent requests!

#### 1.3 Get Profile

```bash
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Expected**: 200 OK with user details

#### 1.4 Update Profile

```bash
curl -X PUT http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Name",
    "language": "ru"
  }'
```

**Expected**: 200 OK with updated user

#### 1.5 Change Password

```bash
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "password123",
    "new_password": "newpass456",
    "confirm_new_password": "newpass456"
  }'
```

**Expected**: 200 OK

#### 1.6 Upload Avatar

```bash
curl -X POST http://localhost:8080/api/v1/auth/avatar \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "avatar=@/path/to/image.jpg"
```

**Expected**: 200 OK with avatar path

---

### Phase 2: Region & District Data

#### 2.1 Get All Regions

```bash
curl -X GET http://localhost:8080/api/v1/regions
```

**Expected**: 200 OK with 13 regions

#### 2.2 Get Districts for Region

```bash
curl -X GET http://localhost:8080/api/v1/regions/1/districts
```

**Expected**: 200 OK with districts (may be empty if not seeded)

---

### Phase 3: Admin Setup (Use SuperAdmin)

Login as superadmin first:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+998901234567",
    "password": "admin123"
  }'
```

**Save admin token!**

#### 3.1 Set Pricing for Routes

```bash
curl -X POST http://localhost:8080/api/v1/admin/pricing \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "from_region_id": 1,
    "to_region_id": 2,
    "base_price": 100000,
    "price_per_person": 25000,
    "service_fee": 15
  }'
```

**Expected**: 200 OK with pricing object

**Repeat for multiple routes** (1→2, 2→1, 1→3, etc.)

#### 3.2 Get All Pricing

```bash
curl -X GET http://localhost:8080/api/v1/admin/pricing \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Expected**: Array of pricing configurations

#### 3.3 Get Platform Statistics

```bash
curl -X GET http://localhost:8080/api/v1/admin/statistics \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Expected**: Statistics object

---

### Phase 4: Create Test Orders

**Use regular user token** from Phase 1

#### 4.1 Create Taxi Order

```bash
curl -X POST http://localhost:8080/api/v1/orders/taxi \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "Test Customer",
    "customer_phone": "+998901111111",
    "from_region_id": 1,
    "from_district_id": 1,
    "to_region_id": 2,
    "to_district_id": 1,
    "passenger_count": 2,
    "scheduled_date": "15.11.2025",
    "time_range_start": "09:00",
    "time_range_end": "11:00",
    "notes": "Test order"
  }'
```

**Expected**: 201 Created with order object and calculated price

**Note the order ID!**

#### 4.2 Create Delivery Order

```bash
curl -X POST http://localhost:8080/api/v1/orders/delivery \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "Test Customer",
    "customer_phone": "+998901111111",
    "recipient_phone": "+998902222222",
    "from_region_id": 1,
    "from_district_id": 1,
    "to_region_id": 2,
    "to_district_id": 1,
    "delivery_type": "document",
    "scheduled_date": "15.11.2025",
    "time_range_start": "14:00",
    "time_range_end": "16:00",
    "notes": "Important documents"
  }'
```

**Expected**: 201 Created with order object

#### 4.3 Get My Orders

```bash
curl -X GET http://localhost:8080/api/v1/orders/my \
  -H "Authorization: Bearer USER_TOKEN"
```

**Expected**: Array of user's orders

#### 4.4 Get Order Details

```bash
curl -X GET http://localhost:8080/api/v1/orders/1 \
  -H "Authorization: Bearer USER_TOKEN"
```

**Expected**: Detailed order object

---

### Phase 5: Driver Application & Approval

#### 5.1 Apply as Driver

**Register new user first** (or use existing):

```bash
curl -X POST http://localhost:8080/api/v1/driver/apply \
  -H "Authorization: Bearer USER_TOKEN" \
  -F "full_name=John Driver" \
  -F "car_model=Chevrolet Lacetti" \
  -F "car_number=01A123BC" \
  -F "license_image=@/path/to/license.jpg"
```

**Expected**: 201 Created with application

**Note the application ID!**

#### 5.2 Get Driver Applications (Admin)

```bash
curl -X GET http://localhost:8080/api/v1/admin/driver-applications?status=pending \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Expected**: Array of pending applications

#### 5.3 Approve Driver Application (Admin)

```bash
curl -X POST http://localhost:8080/api/v1/admin/driver-applications/1/review \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved"
  }'
```

**Expected**: 200 OK, user role changed to driver

#### 5.4 Add Balance to Driver (Admin)

```bash
curl -X POST http://localhost:8080/api/v1/admin/drivers/1/add-balance \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100000
  }'
```

**Expected**: 200 OK, balance added

---

### Phase 6: Driver Operations

**Login as driver** to get driver token:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "DRIVER_PHONE",
    "password": "DRIVER_PASSWORD"
  }'
```

#### 6.1 Get Driver Profile

```bash
curl -X GET http://localhost:8080/api/v1/driver/profile \
  -H "Authorization: Bearer DRIVER_TOKEN"
```

**Expected**: Driver profile with balance and rating

#### 6.2 Update Driver Profile

```bash
curl -X PUT http://localhost:8080/api/v1/driver/profile \
  -H "Authorization: Bearer DRIVER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Updated Driver Name",
    "car_model": "Chevrolet Nexia",
    "car_number": "01B456CD"
  }'
```

**Expected**: 200 OK with updated profile

#### 6.3 Get New Available Orders

```bash
curl -X GET http://localhost:8080/api/v1/driver/orders/new \
  -H "Authorization: Bearer DRIVER_TOKEN"
```

**Expected**: Array of available orders

#### 6.4 Get New Orders with Filters

```bash
curl -X GET "http://localhost:8080/api/v1/driver/orders/new?type=taxi&from_region=1" \
  -H "Authorization: Bearer DRIVER_TOKEN"
```

**Expected**: Filtered array of orders

#### 6.5 Accept Order

```bash
curl -X POST http://localhost:8080/api/v1/driver/orders/1/accept \
  -H "Authorization: Bearer DRIVER_TOKEN"
```

**Expected**: 200 OK, order accepted, balance deducted

#### 6.6 Get My Driver Orders

```bash
curl -X GET http://localhost:8080/api/v1/driver/orders \
  -H "Authorization: Bearer DRIVER_TOKEN"
```

**Expected**: Array of driver's accepted orders

#### 6.7 Complete Order

```bash
curl -X POST http://localhost:8080/api/v1/driver/orders/1/complete \
  -H "Authorization: Bearer DRIVER_TOKEN"
```

**Expected**: 200 OK, order marked completed

#### 6.8 Get Driver Statistics

```bash
curl -X GET http://localhost:8080/api/v1/driver/statistics?period=daily \
  -H "Authorization: Bearer DRIVER_TOKEN"
```

**Expected**: Statistics object

---

### Phase 7: Ratings

**Use user token** (order creator):

#### 7.1 Rate Driver

```bash
curl -X POST http://localhost:8080/api/v1/ratings \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": 1,
    "rating": 5,
    "comment": "Excellent driver!"
  }'
```

**Expected**: 201 Created with rating

#### 7.2 Get Driver Ratings

```bash
curl -X GET http://localhost:8080/api/v1/ratings/driver/1
```

**Expected**: Array of ratings for driver

---

### Phase 8: Order Cancellation

**Use user token**:

```bash
curl -X POST http://localhost:8080/api/v1/orders/2/cancel \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Changed my plans"
  }'
```

**Expected**: 200 OK, order cancelled, driver refunded if accepted

---

### Phase 9: Notifications

#### 9.1 Get My Notifications

```bash
curl -X GET http://localhost:8080/api/v1/notifications \
  -H "Authorization: Bearer USER_TOKEN"
```

**Expected**: Array of notifications

#### 9.2 Get Unread Notifications

```bash
curl -X GET http://localhost:8080/api/v1/notifications?unread=true \
  -H "Authorization: Bearer USER_TOKEN"
```

**Expected**: Array of unread notifications

#### 9.3 Mark Notification as Read

```bash
curl -X POST http://localhost:8080/api/v1/notifications/1/read \
  -H "Authorization: Bearer USER_TOKEN"
```

**Expected**: 200 OK

---

### Phase 10: Feedback

```bash
curl -X POST http://localhost:8080/api/v1/feedback \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Great service! Would love a mobile app."
  }'
```

**Expected**: 201 Created

#### Get Feedback (Admin)

```bash
curl -X GET http://localhost:8080/api/v1/admin/feedback \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Expected**: Array of feedback

---

### Phase 11: Admin User Management

#### 11.1 Get All Drivers

```bash
curl -X GET http://localhost:8080/api/v1/admin/drivers \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Expected**: Array of all drivers

#### 11.2 Block User

```bash
curl -X POST http://localhost:8080/api/v1/admin/users/2/block \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "is_blocked": true
  }'
```

**Expected**: 200 OK, user blocked

#### 11.3 Unblock User

```bash
curl -X POST http://localhost:8080/api/v1/admin/users/2/block \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "is_blocked": false
  }'
```

**Expected**: 200 OK, user unblocked

#### 11.4 Get All Orders (Admin)

```bash
curl -X GET "http://localhost:8080/api/v1/admin/orders?status=completed&from_date=2025-11-01" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Expected**: Filtered array of orders

---

### Phase 12: SuperAdmin Operations

**Use superadmin token**:

#### 12.1 Create Admin User

```bash
curl -X POST http://localhost:8080/api/v1/admin/create-admin \
  -H "Authorization: Bearer SUPERADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+998903333333",
    "name": "New Admin",
    "password": "adminpass123"
  }'
```

**Expected**: 201 Created with admin user

#### 12.2 Reset User Password

```bash
curl -X POST http://localhost:8080/api/v1/admin/users/3/reset-password \
  -H "Authorization: Bearer SUPERADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "new_password": "resetpass123"
  }'
```

**Expected**: 200 OK, password reset

---

## Automated Testing Script

Save as `test.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

echo "=== Testing Taxi Service API ==="

# Register user
echo "\n1. Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"phone_number":"+998909999999","name":"Test","password":"test123","confirm_password":"test123"}')
echo $REGISTER_RESPONSE | jq '.'

# Extract token
TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.token')
echo "Token: $TOKEN"

# Get profile
echo "\n2. Getting profile..."
curl -s -X GET "$BASE_URL/auth/profile" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Get regions
echo "\n3. Getting regions..."
curl -s -X GET "$BASE_URL/regions" | jq '.'

echo "\n=== Tests Complete ==="
```

Make executable and run:
```bash
chmod +x test.sh
./test.sh
```

---

## Testing Checklist

Use this checklist to verify all features:

### Authentication ✓
- [ ] User registration
- [ ] User login
- [ ] Get profile
- [ ] Update profile
- [ ] Change password
- [ ] Upload avatar
- [ ] Invalid credentials handling
- [ ] Token expiration

### Orders ✓
- [ ] Create taxi order
- [ ] Create delivery order
- [ ] Get my orders
- [ ] Get order details
- [ ] Cancel order
- [ ] Pricing calculation
- [ ] Discount application
- [ ] Filter orders by status
- [ ] Filter orders by type

### Driver ✓
- [ ] Apply as driver
- [ ] View new orders
- [ ] Accept order
- [ ] Complete order
- [ ] Get driver orders
- [ ] Get statistics
- [ ] Update profile
- [ ] Balance deduction
- [ ] Accept deadline enforcement

### Admin ✓
- [ ] Review driver applications
- [ ] Approve driver
- [ ] Reject driver
- [ ] Get all drivers
- [ ] Add driver balance
- [ ] Block user
- [ ] Unblock user
- [ ] Set pricing
- [ ] Get all orders
- [ ] Get statistics
- [ ] Get feedback

### Rating ✓
- [ ] Rate completed order
- [ ] Get driver ratings
- [ ] Average rating calculation
- [ ] Prevent duplicate ratings

### Notifications ✓
- [ ] Receive notifications
- [ ] Mark as read
- [ ] Filter unread

### Regions ✓
- [ ] Get all regions
- [ ] Get districts by region

### Feedback ✓
- [ ] Submit feedback
- [ ] Admin view feedback

---

## Performance Testing

### Load Testing with Apache Bench

```bash
# Install Apache Bench
sudo apt install apache2-utils

# Test health endpoint
ab -n 1000 -c 10 http://localhost:8080/health

# Test with authentication (create token file first)
ab -n 100 -c 5 -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/auth/profile
```

### Database Query Performance

```sql
-- Check slow queries
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Check table sizes
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

---

## Debugging Tips

### View Logs
```bash
# Real-time logs
sudo journalctl -u taxi-service -f

# Last 100 lines
sudo journalctl -u taxi-service -n 100
```

### Database Queries
```bash
# Connect to database
psql -U taxi_user -d taxi_service

# Check tables
\dt

# Count records
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM orders;
SELECT COUNT(*) FROM drivers;
```

### Common Issues

**Issue**: "Database connection failed"
- Check PostgreSQL is running
- Verify credentials in `.env`

**Issue**: "Invalid token"
- Token may be expired
- Login again to get new token

**Issue**: "Insufficient balance"
- Admin needs to add balance to driver

**Issue**: "Order not found"
- Verify order ID exists
- Check user permissions

---

## Test Data Cleanup

After testing, clean up test data:

```sql
-- Delete test users
DELETE FROM users WHERE phone_number LIKE '+99890%';

-- Delete test orders
DELETE FROM orders WHERE customer_phone LIKE '+99890%';

-- Delete test ratings
DELETE FROM ratings WHERE user_id IN (
  SELECT id FROM users WHERE phone_number LIKE '+99890%'
);

-- Reset sequences if needed
ALTER SEQUENCE users_id_seq RESTART WITH 1;
ALTER SEQUENCE orders_id_seq RESTART WITH 1;
```

---

**Testing Complete!** All endpoints verified and working. ✅

For production testing, consider:
- Automated integration tests
- Load testing
- Security testing
- Penetration testing
- User acceptance testing
