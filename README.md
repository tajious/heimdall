# Heimdall

A high-performance authentication microservice built with Go and Fiber.

## Features

- Multi-tenant authentication system
- JWT-based authentication with customizable expiration
- Role-based access control
- Multiple authentication methods:
  - Username/Password
- Rate limiting per IP and user
- PostgreSQL support in production
- In-memory storage for development
- Redis caching in production
- Environment-based configuration


## Configuration

The service is configured using environment variables. Create a `.env` file with the following variables:

```env
# Server Configuration
PORT=8080
ENVIRONMENT=development

# Database Configuration
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=heimdall
DB_SSL_MODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your-secret-key
JWT_EXPIRATION_MINUTES=60

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT=100
RATE_LIMIT_WINDOW=60
```

## API Documentation

### Authentication

All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <token>
```

### Endpoints

#### Authentication

##### Login
- **URL**: `POST /api/v1/:tenant_id/login`
- **Description**: Authenticate a user and get a JWT token
- **Rate Limit**: 5 requests per minute
- **Request**:
```json
{
  "username": "string",
  "password": "string",
  "phone": "string" // optional
}
```
- **Response**:
```json
{
  "token": "string",
  "expires_in": 0,
  "user": {
    "id": "string",
    "tenant_id": "string",
    "username": "string",
    "phone": "string",
    "role": "string",
    "last_login": "string",
    "created_at": "string",
    "updated_at": "string"
  }
}
```

##### Validate Token
- **URL**: `POST /api/v1/validate-token`
- **Description**: Validate a JWT token
- **Request**:
```json
{
  "token": "string"
}
```
- **Response**:
```json
{
  "valid": true,
  "user": {
    "id": "string",
    "username": "string",
    "role": "string"
  },
  "tenant": {
    "id": "string",
    "name": "string",
    "config": {
      "id": "string",
      "tenant_id": "string",
      "auth_method": "string",
      "jwt_duration": 0,
      "rate_limit_ip": 0,
      "rate_limit_user": 0,
      "rate_limit_window": 0,
      "created_at": "string",
      "updated_at": "string"
    }
  },
  "expires_at": "string"
}
```

#### Tenants

##### Create Tenant
- **URL**: `POST /api/v1/tenants`
- **Description**: Create a new tenant
- **Request**:
```json
{
  "name": "string",
  "description": "string",
  "auth_method": "username_password",
  "jwt_duration": 0,
  "rate_limit_ip": 0,
  "rate_limit_user": 0,
  "rate_limit_window": 0
}
```
- **Response**:
```json
{
  "id": "string",
  "name": "string",
  "config": {
    "id": "string",
    "tenant_id": "string",
    "auth_method": "string",
    "jwt_duration": 0,
    "rate_limit_ip": 0,
    "rate_limit_user": 0,
    "rate_limit_window": 0,
    "created_at": "string",
    "updated_at": "string"
  },
  "created_at": "string",
  "updated_at": "string"
}
```

##### Get Tenant
- **URL**: `GET /api/v1/tenants/:tenant_id`
- **Description**: Get a single tenant by ID
- **Authentication**: Required
- **Response**:
```json
{
  "id": "string",
  "name": "string",
  "config": {
    "id": "string",
    "tenant_id": "string",
    "auth_method": "string",
    "jwt_duration": 0,
    "rate_limit_ip": 0,
    "rate_limit_user": 0,
    "rate_limit_window": 0,
    "created_at": "string",
    "updated_at": "string"
  },
  "created_at": "string",
  "updated_at": "string"
}
```

##### List Tenants
- **URL**: `GET /api/v1/tenants`
- **Description**: List all tenants with pagination
- **Authentication**: Required
- **Query Parameters**:
  - `page` (optional, default: 1): Page number
  - `page_size` (optional, default: 10): Number of items per page
- **Response**:
```json
{
  "tenants": [
    {
      "id": "string",
      "name": "string",
      "config": {
        "id": "string",
        "tenant_id": "string",
        "auth_method": "string",
        "jwt_duration": 0,
        "rate_limit_ip": 0,
        "rate_limit_user": 0,
        "rate_limit_window": 0,
        "created_at": "string",
        "updated_at": "string"
      },
      "created_at": "string",
      "updated_at": "string"
    }
  ],
  "total": 0,
  "page": 0,
  "page_size": 0,
  "total_pages": 0
}
```

##### Update Tenant Config
- **URL**: `PUT /api/v1/tenants/:tenant_id/config`
- **Description**: Update tenant configuration
- **Authentication**: Required
- **Request**:
```json
{
  "auth_method": "username_password",
  "jwt_duration": 0,
  "rate_limit_ip": 0,
  "rate_limit_user": 0,
  "rate_limit_window": 0
}
```
- **Response**:
```json
{
  "message": "Tenant configuration updated successfully",
  "config": {
    "id": "string",
    "tenant_id": "string",
    "auth_method": "string",
    "jwt_duration": 0,
    "rate_limit_ip": 0,
    "rate_limit_user": 0,
    "rate_limit_window": 0,
    "created_at": "string",
    "updated_at": "string"
  }
}
```

#### Users

##### List Users
- **URL**: `GET /api/v1/tenants/:tenant_id/users`
- **Description**: List users for a tenant with pagination
- **Authentication**: Required
- **Query Parameters**:
  - `page` (optional, default: 1): Page number
  - `page_size` (optional, default: 10): Number of items per page
  - `search` (optional): Search term for username or phone
  - `role` (optional): Filter by role
  - `sort_by` (optional): Sort field (username, role, created_at, last_login)
  - `sort_dir` (optional): Sort direction (asc, desc)
- **Response**:
```json
{
  "users": [
    {
      "id": "string",
      "tenant_id": "string",
      "username": "string",
      "phone": "string",
      "role": "string",
      "last_login": "string",
      "created_at": "string",
      "updated_at": "string"
    }
  ],
  "total": 0,
  "page": 0,
  "page_size": 0,
  "total_pages": 0
}
```

##### Get Current User
- **URL**: `GET /api/v1/me`
- **Description**: Get current user information
- **Authentication**: Required
- **Response**: JWT claims of the current user

## Development

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Create a `.env` file with your configuration. For development, you only need these variables:
   ```env
   # Server Configuration
   PORT=8080
   ENVIRONMENT=development

   # JWT Configuration
   JWT_SECRET=your-secret-key
   JWT_EXPIRATION_MINUTES=60

   # Rate Limiting
   RATE_LIMIT_ENABLED=true
   RATE_LIMIT=100
   RATE_LIMIT_WINDOW=60
   ```
   Note: In development environment, the service uses in-memory storage, so you don't need to configure PostgreSQL or Redis.

4. Run the server:
   ```bash
   go run cmd/main.go
   ```

## Production

1. Build the application:
   ```bash
   go build -o heimdall cmd/main.go
   ```
2. Set up PostgreSQL and Redis
3. Configure environment variables. For production, you need all variables:
   ```env
   # Server Configuration
   PORT=8080
   ENVIRONMENT=production

   # Database Configuration
   DB_DRIVER=postgres
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=heimdall
   DB_SSL_MODE=disable

   # Redis Configuration
   REDIS_HOST=localhost
   REDIS_PORT=6379
   REDIS_PASSWORD=
   REDIS_DB=0

   # JWT Configuration
   JWT_SECRET=your-secret-key
   JWT_EXPIRATION_MINUTES=60

   # Rate Limiting
   RATE_LIMIT_ENABLED=true
   RATE_LIMIT=100
   RATE_LIMIT_WINDOW=60
   ```
4. Run the application:
   ```bash
   ./heimdall
   ```

## License

MIT 