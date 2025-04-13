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

## Project Structure

```
heimdall/
├── cmd/
│   └── heimdall/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handlers/
│   │   └── auth.go
│   ├── middleware/
│   │   ├── auth.go
│   │   └── rate_limiter.go
│   ├── models/
│   │   ├── tenant.go
│   │   └── user.go
│   └── storage/
│       └── storage.go
├── pkg/
│   ├── auth/
│   └── rate_limiter/
├── .env
└── README.md
```

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

## API Endpoints

### Public Endpoints

#### Create Tenant
- `POST /api/v1/tenants`
  - Creates a new tenant with its configuration
  - Request body:
    ```json
    {
      "name": "string",
      "description": "string",
      "auth_method": "username_password",
      "jwt_duration": 3600,
      "rate_limit_ip": 100,
      "rate_limit_user": 50,
      "rate_limit_window": 60
    }
    ```
  - Response:
    ```json
    {
      "id": "string",
      "name": "string",
      "description": "string",
      "config": {
        "auth_method": "username_password",
        "jwt_duration": 3600,
        "rate_limit_ip": 100,
        "rate_limit_user": 50,
        "rate_limit_window": 60,
        "created_at": "2024-04-12T12:00:00Z",
        "updated_at": "2024-04-12T12:00:00Z"
      }
    }
    ```

#### Login
- `POST /api/v1/:tenant_id/login`
  - Authenticates a user with the specified tenant
  - Rate Limited: Yes (5 requests per minute)
  - Request body:
    ```json
    {
      "username": "string",
      "password": "string"
    }
    ```
  - Response:
    ```json
    {
      "token": "string",
      "expires_in": 3600,
      "user": {
        "id": "string",
        "tenant_id": "string",
        "username": "string",
        "role": "string",
        "last_login": "2024-04-12T12:00:00Z",
        "created_at": "2024-04-12T12:00:00Z",
        "updated_at": "2024-04-12T12:00:00Z"
      }
    }
    ```

#### Validate Token
- `POST /api/v1/validate-token`
  - Validates a JWT token and returns user/tenant information
  - Headers:
    ```
    Authorization: Bearer <token>
    ```
  - Response:
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
          "auth_method": "username_password",
          "jwt_duration": 3600,
          "rate_limit_ip": 100,
          "rate_limit_user": 50,
          "rate_limit_window": 60
        }
      },
      "expires_at": "2024-04-12T12:00:00Z"
    }
    ```

### Protected Endpoints (Requires Authentication)

#### Get Current User
- `GET /api/v1/me`
  - Returns information about the currently authenticated user
  - Headers:
    ```
    Authorization: Bearer <token>
    ```
  - Response:
    ```json
    {
      "id": "string",
      "tenant_id": "string",
      "username": "string",
      "role": "string",
      "last_login": "2024-04-12T12:00:00Z",
      "created_at": "2024-04-12T12:00:00Z",
      "updated_at": "2024-04-12T12:00:00Z"
    }
    ```

#### Update Tenant Configuration
- `PUT /api/v1/tenants/:tenant_id/config`
  - Updates the configuration for a specific tenant
  - Headers:
    ```
    Authorization: Bearer <token>
    ```
  - Request body:
    ```json
    {
      "auth_method": "username_password",
      "jwt_duration": 3600,
      "rate_limit_ip": 100,
      "rate_limit_user": 50,
      "rate_limit_window": 60
    }
    ```
  - Response:
    ```json
    {
      "message": "Tenant configuration updated successfully",
      "config": {
        "auth_method": "username_password",
        "jwt_duration": 3600,
        "rate_limit_ip": 100,
        "rate_limit_user": 50,
        "rate_limit_window": 60,
        "created_at": "2024-04-12T12:00:00Z",
        "updated_at": "2024-04-12T12:30:00Z"
      }
    }
    ```

#### List Users
- `GET /api/v1/tenants/:tenant_id/users`
  - Lists users with pagination, search, filtering, and sorting
  - Headers:
    ```
    Authorization: Bearer <token>
    ```
  - Query Parameters:
    - `page`: Page number (default: 1)
    - `page_size`: Number of items per page (default: 10, max: 100)
    - `search`: Search term for username or phone
    - `role`: Filter by user role
    - `sort_by`: Field to sort by (username, role, created_at, last_login)
    - `sort_dir`: Sort direction (asc, desc)
  - Response:
    ```json
    {
      "users": [
        {
          "id": "string",
          "tenant_id": "string",
          "username": "string",
          "role": "string",
          "last_login": "2024-04-12T12:00:00Z",
          "created_at": "2024-04-12T12:00:00Z",
          "updated_at": "2024-04-12T12:00:00Z"
        }
      ],
      "total": 100,
      "page": 1,
      "page_size": 10,
      "total_pages": 10
    }
    ```

## Development

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Create a `.env` file with your configuration
4. Run the server:
   ```bash
   go run cmd/heimdall/main.go
   ```

## Production

1. Build the application:
   ```bash
   go build -o heimdall cmd/heimdall/main.go
   ```
2. Set up PostgreSQL and Redis
3. Configure environment variables
4. Run the application:
   ```bash
   ./heimdall
   ```

## License

MIT 