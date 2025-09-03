# Chirpy RESTful API Documentation

Chirpy is a microblogging platform API allowing users to create accounts, log in, post chirps, and manage their account.

**Base URL:** `http://localhost:8080`

**Authentication:**  
- User endpoints require **JWT Bearer tokens** (Authorization header).  
- Polka webhook endpoint requires **Polka API key** in headers.  

***
## Table of Contents

1. [User Endpoints](#user-endpoints)  
2. [Authentication Endpoints](#authentication-endpoints)  
3. [Chirp Endpoints](#chirp-endpoints)  
4. [Health Check Endpoint](#health-check-endpoint)  
5. [Admin/Metric Endpoints](#adminmetric-endpoints)  
6. [Static File Endpoints](#static-file-endpoints)  

***
## User Endpoints

### Create User

```
- Endpoint: `POST /api/users`  
- Description: Register a new user.  
- Authentication: None  
```

**Request Body:**

```
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Responses:**
- `201 Created` – Returns created user object:

```
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

- `400 Bad Request` – Invalid email/password.  
- `409 Conflict` – User already exists.  
- `500 Internal Server Error` – Failed to create user.

***
### Update User

- Endpoint: `PUT /api/users`  
- Description: Update the email or password of the authenticated user.  
- Authentication: JWT Bearer token required (`Authorization: Bearer <token>`).  

**Request Body:**
```
{
  "email": "newemail@example.com",
  "password": "newpassword123"
}
```

**Responses:**
- `200 OK` – Updated user object returned.  
- `400 Bad Request` – Missing fields or invalid token.  
- `500 Internal Server Error` – Database or server error.

***
## Authentication Endpoints

### Login

- Endpoint: `POST /api/login`  
- Description: Authenticate a user and return JWT + refresh token.  
- Request Body:

```
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Responses:**
- `200 OK` – Returns user info with tokens:

```
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "email": "user@example.com",
  "is_chirpy_red": false,
  "token": "jwt_token_here",
  "refresh_token": "refresh_token_here"
}
```

- `401 Unauthorized` – Incorrect email/password.  
- `500 Internal Server Error` – Server failure.

### Refresh Token

- Endpoint: `POST /api/refresh`  
- Description: Refresh JWT using a valid refresh token.  
- Authentication: Authorization header with **refresh token** (`Authorization: Bearer <refresh_token>`).  

**Responses:**
- `200 OK` – Returns new JWT:

```
{
  "token": "new_jwt_token_here"
}
```

- `401 Unauthorized` – Invalid/expired refresh token.  
- `500 Internal Server Error` – Server failure.

***
### Revoke Refresh Token

- Endpoint: `POST /api/revoke`  
- Description: Revoke a refresh token.  
- Authentication: Authorization header with **refresh token**.  

**Responses:**
- `204 No Content` – Successfully revoked.  
- `400 Bad Request` – Invalid token.  
- `500 Internal Server Error` – Server failure.

***
### Polka Webhook: Upgrade User

- Endpoint: `POST /api/polka/webhooks`  
- Description: Upgrade a user to Chirpy Red via Polka webhook.  
- Authentication: Polka API key (`X-API-Key: <key>`).  
- Request Body:

```
{
  "event": "user.upgraded",
  "data": {
    "user_id": "uuid"
  }
}
```

**Responses:**
- `204 No Content` – Success or ignored event.  
- `400 Bad Request` – Invalid API key or body.  
- `404 Not Found` – User not found.  
- `500 Internal Server Error` – Server failure.

***
## Chirp Endpoints

### Create Chirp

- Endpoint: `POST /api/chirps`  
- Description: Create a new chirp (max 140 characters).  
- Authentication: JWT Bearer token required.  
- Request Body:

```
{
  "body": "Hello, this is my first chirp!"
}
```

**Responses:**
- `201 Created` – Returns the created chirp:

```
{
  "id": "uuid",
  "body": "Hello, this is my first chirp!",
  "user_id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

- `400 Bad Request` – Empty or too long chirp, or invalid token.  
- `500 Internal Server Error` – Server failure.

***
### Get All Chirps

- Endpoint: `GET /api/chirps`  
- Description: Retrieve all chirps or filter by author.  
- Query Parameters:
  - `author_id` (optional) – UUID of author.  
  - `sort` (optional) – `"desc"` for descending order by creation date.  

**Responses:**
- `200 OK` – Returns array of chirps.  
- `404 Not Found` – No chirps found.  
- `500 Internal Server Error` – Server failure.

***
### Get One Chirp

- Endpoint: `GET /api/chirps/{chirpID}`  
- Description: Retrieve a single chirp by ID.  
- Path Parameter: `chirpID` – UUID of the chirp.  

**Responses:**
- `200 OK` – Returns the chirp object.  
- `400 Bad Request` – Invalid `chirpID`.  
- `404 Not Found` – Chirp not found.  
- `500 Internal Server Error` – Server failure.

***
### Delete Chirp

- Endpoint: `DELETE /api/chirps/{chirpID}`  
- Description: Delete a chirp by ID.  
- Authentication: JWT Bearer token required.  
- Path Parameter: `chirpID` – UUID of the chirp.  

**Responses:**
- `204 No Content` – Successfully deleted.  
- `400 Bad Request` – Invalid `chirpID`.  
- `404 Not Found` – Chirp not found.  
- `500 Internal Server Error` – Server failure.

***
## Health Check Endpoint

### Health

- Endpoint: `GET /api/healthz`  
- Description: Check if the API is running.  
**Responses:**  
- `200 OK` – Healthy status.

***
## Admin/Metric Endpoints

### Metrics

- Endpoint: `GET /admin/metrics`  
- Description: Show metrics for file server hits.  
**Responses:**
- `200 OK` – Returns HTML page with hits count.

***
### Reset

- Endpoint: `POST /admin/reset`  
- Description: Reset metrics and delete all users (dev only).  
- Authentication: None, but restricted to `dev` platform.  
**Responses:**
- `200 OK` – Reset successful.  
- `403 Forbidden` – Not allowed on non-dev platform.  
- `500 Internal Server Error` – Failure during reset.

***
## Static File Endpoints

- Serve App UI: `/app/` → serves static files from `./static/`  
- Logo: `/app/logo.png` → serves `./static/assets/logo.png`  
- Authentication: None  
- Metrics: File hits are tracked via middleware.

***
## Authentication Notes

- JWT Token: Sent as `Authorization: Bearer <token>` in headers.  
- Refresh Token: Used for `/api/refresh` and `/api/revoke`.  
- Polka Key: Sent as `X-API-Key` in headers for webhook upgrade.
***

