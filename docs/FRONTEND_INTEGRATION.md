# Frontend Integration Guide

This document describes how the frontend clients should communicate with the Fiber-based Taxi Service API.

## Base URLs

- Production API: `https://api.omad-driver.uz/api/v1`
- Production Swagger: `https://api.omad-driver.uz/swagger/index.html`
- Health check: `https://api.omad-driver.uz/health`

Use `http://localhost:8080` during local development unless otherwise configured in `.env`.

## Authentication

- Login endpoint: `POST /auth/login`
- Returns:
  ```json
  {
    "token": "<jwt>",
    "role": "admin",
    "user": { ... }
  }
  ```
- Persist the JWT (e.g. secure storage); include it in the `Authorization` header for all protected requests:
  ```http
  Authorization: Bearer <jwt>
  ```
- Token payload (`claims`): `user_id`, `role`, `exp`, `iat`.

## Roles & Routing

- `user` – regular consumers (rides & deliveries).
- `driver` – approved drivers; they also have access to `driver` endpoints.
- `admin` – admin console access.
- `superadmin` – everything admins can do plus admin management.

Use the `role` returned from `/auth/login` to route users to the appropriate UI surfaces.

## New/Updated Endpoints

### Auth

- `POST /auth/register` – unchanged.
- `POST /auth/login` – now returns `role` alongside the JWT.
- `GET /auth/profile` / `PUT /auth/profile` – unchanged.
- `POST /auth/avatar` – multipart upload; max size 10MB. Upload URL is returned in response.

### Drivers

- `POST /driver/apply` – multipart form (`full_name`, `car_model`, `car_number`, `license_image`).
- `GET /driver/orders/new?type=taxi&from_region=1&to_region=2` – filters are optional.
- `POST /driver/orders/:id/accept` – will return `409` if another driver already claimed the order.
- `POST /driver/orders/:id/complete` – marks order as completed.

### Admin Area

- `POST /admin/driver-applications/:id/review`
  ```json
  {
    "status": "approved", // or "rejected"
    "rejection_reason": "Optional when rejected"
  }
  ```
- `POST /admin/drivers/:id/add-balance`
  ```json
  { "amount": 50000 }
  ```
- `POST /admin/pricing`
  ```json
  {
    "from_region_id": 1,
    "to_region_id": 5,
    "base_price": 150000,
    "price_per_person": 25000,
    "service_fee": 15
  }
  ```

### Regions & Districts

- `GET /regions` – returns all 14 Uzbekistan regions.
- `GET /regions/:id/districts` – returns districts filtered by region.
- `GET /districts/:id` – fetch a single district.

### Pricing

- `GET /admin/pricing` – list of routes and associated pricing.

## Reference Data

A dedicated seeding tool (`make db-seed`) loads all Uzbekistan regions, districts, and realistic pricing combinations. The API always responds with IDs referencing this data.

## Error Handling

All errors are JSON objects with an `error` field and HTTP status codes:

```json
{
  "error": "Invalid or expired token"
}
```

Important statuses:

- `401` – missing/invalid token.
- `403` – insufficient role permissions.
- `409` – conflicting actions (e.g. duplicate application or order already taken).

## CORS & Origins

Allowed production origins:

- `https://api.omad-driver.uz`
- `https://docs.omad-driver.uz`
- Add more comma-separated domains via `.env` (`CORS_ALLOWED_ORIGINS`).

## Swagger Usage

Swagger lives at `/swagger/index.html` and reflects the latest Fiber routes. Use it for up-to-date request/response schemas.

## Frontend Checklist

- Store and forward JWT tokens with `Authorization: Bearer` header.
- Handle role-specific routing after login.
- Respect the multipart requirements for uploads (`avatar`, `license_image`).
- Use `/regions` & `/districts` endpoints to populate selection lists (IDs, names).
- Surface descriptive error messages from the API (use response `error` string).
- Poll or refresh order lists for drivers regularly to capture newly available rides.

