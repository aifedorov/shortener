# URL Shortener Service

## 🚀 What This Service Does

This service provides a RESTful API for shortening long URLs into compact. It supports:

- **Single URL shortening** via plain text and JSON endpoints
- **Batch URL processing** for multiple URLs at once
- **User authentication** with JWT-based session management
- **URL management** - users can view and delete their shortened URLs
- **Health monitoring** with ping endpoints
- **Multiple storage backends** - in-memory, file-based, and PostgreSQL

## 🛠 Technologies & Libraries Used

### Core Technologies
- **Go 1.21+** - Primary programming language
- **Chi Router** - Lightweight, idiomatic HTTP router
- **PostgreSQL** - Primary database
- **JWT** - JSON Web Tokens for authentication
- **Gzip** - Compression for response optimization

### Key Libraries & Frameworks
- **`github.com/go-chi/chi/v5`** - HTTP router and middleware
- **`github.com/golang-jwt/jwt/v4`** - JWT token handling
- **`github.com/jackc/pgx/v5`** - PostgreSQL driver
- **`go.uber.org/zap`** - Structured logging
- **`github.com/google/uuid`** - UUID generation
- **`github.com/stretchr/testify`** - Testing framework
- **`github.com/golang/mock/gomock`** - Mock generation

### Architecture Patterns
- **Clean Architecture** - Separation of concerns with distinct layers
- **Repository Pattern** - Abstracted data access layer
- **Middleware Pattern** - Cross-cutting concerns (auth, logging, compression)

## 📁 Project Structure

```
├── cmd/shortener/          # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── http/              # HTTP layer
│   │   ├── handlers/      # Request handlers
│   │   └── middleware/    # HTTP middleware (auth, logging, compression)
│   ├── pkg/               # Internal packages
│   │   ├── random/        # Random string generation
│   │   └── validate/      # URL validation
│   ├── repository/        # Data access layer
│   └── mocks/             # Generated mocks for testing
└── profiles/              # Performance profiling data
```

## 🔧 Key Features Demonstrated

### 1. **Comprehensive Testing Strategy**
- **Unit Tests** - 88%+ coverage for handlers and business logic
- **Integration Tests** - End-to-end workflow testing
- **Mock-based Testing** - Using `gomock` for dependency isolation
- **Example Tests** - Runnable documentation examples

### 2. **Middleware**
- **Authentication** - JWT-based user session management
- **Logging** - Structured logging with request/response tracking
- **Compression** - Gzip compression for response optimization
- **Error Handling** - Comprehensive error responses

### 3. **Flexible Storage Architecture**
- **In-Memory** - Fast development and testing
- **File-based** - Persistent storage without database
- **PostgreSQL** - Production-ready with ACID compliance
- **Repository Pattern** - Easy storage backend switching

### 4. **API Design**
- **RESTful Endpoints** - Standard HTTP methods and status codes
- **Content Negotiation** - Support for JSON and plain text
- **Batch Operations** - Efficient bulk processing

## 🚦 API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/` | Shorten URL (plain text) | ✅ |
| `POST` | `/api/shorten` | Shorten URL (JSON) | ✅ |
| `POST` | `/api/shorten/batch` | Batch URL shortening | ✅ |
| `GET` | `/{shortURL}` | Redirect to original URL | ❌ |
| `GET` | `/api/user/urls` | Get user's URLs | ✅ |
| `DELETE` | `/api/user/urls` | Delete user's URLs | ✅ |
| `GET` | `/ping` | Health check | ❌ |

## 🏃‍♂️ Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL (optional)

### Running the Service

1. **Clone and setup:**
   ```bash
   git clone <repository-url>
   cd shortener
   go mod download
   ```

2. **Set environment variables:**
   ```bash
   export SECRET_KEY="your-secret-key"
   export SERVER_ADDRESS=":8080"
   export BASE_URL="http://localhost:8080"
   ```

3. **Run with in-memory storage:**
   ```bash
   go run cmd/shortener/main.go
   ```

4. **Run with PostgreSQL:**
   ```bash
   export DATABASE_DSN="postgres://user:password@localhost/dbname"
   go run cmd/shortener/main.go
   ```

### Testing
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
```

## 🎯 Skills Demonstrated

This project showcases proficiency in:
- **Go Language** - Idiomatic Go code with proper error handling
- **Web Development** - HTTP servers, middleware, and REST APIs
- **Database Design** - SQL schema design and query optimization
- **Testing** - Comprehensive test coverage with mocks and integration tests
- **DevOps** - Configuration management and deployment considerations
- **Software Architecture** - Clean architecture and design patterns
- **Performance** - Optimization techniques and profiling
- **Documentation** - Comprehensive API documentation and examples
