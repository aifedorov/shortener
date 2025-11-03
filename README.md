# URL Shortener Service

> **Educational Project**: Production-grade URL shortener service built to demonstrate modern Go development practices
> and microservices architecture. Developed as part of a structured learning program with iterative feature development.

## 🚀 What This Service Does

A comprehensive URL shortening service exposing both **REST API** and **gRPC API** for shortening long URLs into
compact, shareable links. Features include:

- **Dual API Support** - HTTP REST and gRPC endpoints for flexibility
- **Single URL shortening** via plain text and JSON endpoints
- **Batch URL processing** for multiple URLs at once
- **User authentication** with JWT-based session management
- **URL management** - users can view and delete their shortened URLs
- **Health monitoring** with ping endpoints and database health checks
- **Multiple storage backends** - in-memory, file-based, and PostgreSQL
- **Production-ready features** - TLS support, graceful shutdown, structured logging

## 🛠 Technologies & Libraries Used

### Core Technologies

- **Go 1.24** - Primary programming language with latest features
- **gRPC + Protocol Buffers** - High-performance RPC framework
- **Chi Router** - Lightweight, idiomatic HTTP router
- **PostgreSQL** - ACID-compliant relational database
- **JWT (HS256)** - Stateless authentication
- **Buf** - Modern Protocol Buffer toolchain

### Key Libraries & Frameworks

- **`google.golang.org/grpc`** - gRPC server and client implementation
- **`google.golang.org/protobuf`** - Protocol Buffers code generation
- **`github.com/go-chi/chi/v5`** - HTTP router and middleware
- **`github.com/golang-jwt/jwt/v4`** - JWT token handling
- **`github.com/jackc/pgx/v5`** - High-performance PostgreSQL driver
- **`go.uber.org/zap`** - Structured, leveled logging
- **`github.com/google/uuid`** - RFC 4122 UUID generation
- **`github.com/stretchr/testify`** - Testing assertions and suites
- **`github.com/golang/mock/gomock`** - Interface mocking for unit tests

### Architecture & Design Patterns

- **Clean Architecture** - Separation of concerns with distinct layers (handlers, business logic, data access)
- **Repository Pattern** - Abstracted data access with multiple backend implementations
- **Middleware Pattern** - Composable cross-cutting concerns (auth, logging, compression)
- **Interceptor Pattern** - gRPC middleware for authentication and logging
- **Dependency Injection** - Testable components with interface-based design

## 📁 Project Structure

```
├── cmd/
│   ├── shortener/          # Main application entry point
│   └── staticlint/         # Custom static analysis tool
├── api/
│   └── grpc/
│       ├── proto/          # Protocol Buffer definitions
│       └── gen/            # Generated gRPC code
├── internal/
│   ├── config/            # Configuration management (flags, env vars, JSON)
│   ├── grpc/              # gRPC server implementation
│   │   ├── middleware/    # gRPC interceptors (auth, logging)
│   │   └── server.go      # gRPC service handlers
│   ├── http/              # HTTP/REST layer
│   │   ├── handlers/      # HTTP request handlers
│   │   ├── middleware/    # HTTP middleware (auth, logging, compression)
│   │   └── server.go      # HTTP server setup
│   ├── pkg/               # Internal shared packages
│   │   ├── jwt/           # JWT generation and validation
│   │   ├── random/        # Random string generation
│   │   └── validate/      # URL validation
│   ├── repository/        # Data access layer (memory, file, PostgreSQL)
│   └── mocks/             # Generated mocks for testing
├── migrations/            # Database schema migrations
└── profiles/              # Performance profiling data
```

## 🔧 Key Features Demonstrated

### 1. **Dual Protocol API Design**

- **REST API** - Standard HTTP/JSON endpoints with Chi router
- **gRPC API** - High-performance binary protocol with Protocol Buffers
- **Unified Business Logic** - Both APIs share the same repository layer
- **Authentication** - JWT tokens via cookies (HTTP) and metadata (gRPC)
- **Content Negotiation** - Support for JSON, plain text, and Protocol Buffers

### 2. **Comprehensive Testing Strategy**
- **Unit Tests** - 88%+ coverage for handlers and business logic
- **Integration Tests** - End-to-end workflow testing for both HTTP and gRPC
- **Mock-based Testing** - Using `gomock` for dependency isolation
- **Example Tests** - Runnable documentation examples
- **Table-Driven Tests** - Idiomatic Go testing patterns

### 3. **Production-Ready Middleware & Interceptors**

- **Authentication** - JWT-based user session management for both HTTP and gRPC
- **Logging** - Structured logging with zap (request/response tracking)
- **Compression** - Gzip compression for HTTP responses
- **Error Handling** - Standardized error responses with proper status codes
- **Public Method Whitelisting** - Secure authentication bypass for health checks

### 4. **Flexible Storage Architecture**
- **In-Memory** - Fast development and testing
- **File-based** - JSON persistent storage without database
- **PostgreSQL** - Production-ready with migrations, connection pooling, ACID compliance
- **Repository Pattern** - Swap backends without changing business logic
- **Automatic Backend Selection** - Based on configuration priority

### 5. **Microservices Best Practices**

- **Configuration Management** - Three-tier priority (flags > env > config file)
- **Graceful Shutdown** - Signal handling for zero-downtime deployments
- **Database Migrations** - Version-controlled schema evolution with golang-migrate
- **Health Checks** - Database connectivity validation
- **TLS Support** - HTTPS with certificate management
- **Structured Logging** - JSON logs with contextual information

## 🚦 API Endpoints

### HTTP REST API (default port: 8080)

| Method   | Endpoint              | Description              | Auth Required     |
|----------|-----------------------|--------------------------|-------------------|
| `POST`   | `/`                   | Shorten URL (plain text) | ✅                 |
| `POST`   | `/api/shorten`        | Shorten URL (JSON)       | ✅                 |
| `POST`   | `/api/shorten/batch`  | Batch URL shortening     | ✅                 |
| `GET`    | `/{shortURL}`         | Redirect to original URL | ❌                 |
| `GET`    | `/api/user/urls`      | Get user's URLs          | ✅                 |
| `DELETE` | `/api/user/urls`      | Delete user's URLs       | ✅                 |
| `GET`    | `/ping`               | Health check             | ❌                 |
| `GET`    | `/api/internal/stats` | Service statistics       | ❌ (IP-restricted) |

### gRPC API (default port: 9090)

| Method                | Description                   | Auth Required     |
|-----------------------|-------------------------------|-------------------|
| `CreateShortURL`      | Create single short URL       | ✅                 |
| `BatchCreateShortURL` | Create multiple short URLs    | ✅                 |
| `GetShortURL`         | Get original URL by short URL | ❌                 |
| `ListShortURLs`       | List user's URLs              | ✅                 |
| `DeleteURLs`          | Soft-delete URLs              | ✅                 |
| `Ping`                | Health check                  | ❌                 |
| `GetStats`            | Service statistics            | ❌ (IP-restricted) |

**Authentication**: JWT tokens passed via `Authorization` cookie (HTTP) or `jwt` metadata (gRPC). Tokens are
auto-generated if not provided.

See [gRPC API Documentation](api/grpc/proto/README.md) for detailed usage examples.

## 🏃‍♂️ Quick Start

### Prerequisites

- Go 1.24+
- PostgreSQL (optional, recommended for production)
- Docker & Docker Compose (optional, for containerized setup)
- Buf CLI (for Protocol Buffer code generation)

### Running the Service

1. **Clone and setup:**
   ```bash
   git clone <repository-url>
   cd shortener
   go mod download
   ```

2. **Set required environment variables:**
   ```bash
   export SECRET_KEY="your-secret-key-minimum-32-characters"
   export SERVER_ADDRESS=":8080"      # HTTP server
   export GRPC_ADDRESS=":9090"        # gRPC server
   export BASE_URL="http://localhost:8080"
   ```

3. **Run with in-memory storage (development):**
   ```bash
   go run cmd/shortener/main.go
   ```

4. **Run with PostgreSQL (recommended):**
   ```bash
   # Start PostgreSQL with Docker
   make docker-db-up

   # Set database connection
   export DATABASE_DSN="postgres://postgres:shortener@localhost:5432/shortener?sslmode=disable"

   # Run migrations
   make migrate-up

   # Start service
   go run cmd/shortener/main.go
   ```

5. **Run with Docker Compose (full stack):**
   ```bash
   make docker-up
   ```

### Testing the APIs

**HTTP REST API:**

```bash
# Shorten a URL
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'

# Health check
curl http://localhost:8080/ping
```

**gRPC API:**

```bash
# Install grpcurl
brew install grpcurl

# Health check
grpcurl -plaintext localhost:9090 shortener.v1.ShortenerService/Ping

# Create short URL
grpcurl -plaintext \
  -d '{"url": "https://example.com"}' \
  localhost:9090 shortener.v1.ShortenerService/CreateShortURL
```

### Running Tests
```bash
# Run all tests
make test
# or
go test ./...

# Run with coverage report
go test -cover ./...

# Run integration tests
go test ./internal/grpc/... -tags=integration
```

### Development Commands

```bash
# Build the application
make build

# Run linter and static analysis
make lint

# Generate gRPC code from proto files
cd api/grpc && buf generate

# Generate mocks for testing
go generate ./...
```

## 🎯 Engineering & Leadership Skills Demonstrated

### Technical Proficiency (1+ year Go experience)

- **Go Language Expertise** - Idiomatic Go 1.24 with generics, proper error handling, and concurrency patterns
- **Microservices Architecture** - Dual-protocol API design (REST + gRPC) with shared business logic
- **API Design** - RESTful principles, gRPC best practices, and Protocol Buffers
- **Database Engineering** - PostgreSQL with migrations, connection pooling, transactions, and query optimization
- **Authentication & Security** - JWT (HS256), stateless sessions, IP-based access control, TLS/HTTPS
- **Testing Excellence** - 88%+ coverage with unit tests, integration tests, mocks, and table-driven tests
- **DevOps & Infrastructure** - Docker, docker-compose, graceful shutdown, configuration management
- **Performance Engineering** - Profiling, optimization, compression, and connection pooling

### Software Engineering Best Practices

- **Clean Architecture** - Separation of concerns with distinct layers (presentation, business, data)
- **Design Patterns** - Repository, middleware, interceptor, dependency injection
- **Code Quality** - Custom static analysis tools, linting, code generation (`gomock`, `protoc`)
- **Documentation** - Comprehensive inline docs, API guides, README, architecture documentation
- **Version Control** - Git workflow with iterative feature branches
- **CI/CD** - Automated testing and static analysis via GitHub Actions

### Engineering Management Capabilities

- **Project Organization** - Clear structure, modular design, and maintainable codebase
- **Technical Decision Making** - Evaluated and implemented appropriate technologies for requirements
- **Iterative Development** - Structured learning program with incremental feature delivery (25+ iterations)
- **Code Review Mindset** - Focus on maintainability, testability, and production readiness
- **System Design** - Designed for scalability, observability, and operational excellence
- **Documentation Culture** - Created comprehensive documentation for team onboarding and API usage

### Specific Technologies

- **Go Ecosystem**: chi, pgx, zap, gomock, testify, jwt-go
- **gRPC Stack**: Protocol Buffers, buf, grpcurl, gRPC interceptors
- **Infrastructure**: PostgreSQL, Docker, golang-migrate
- **Tools**: Make, git, static analyzers

---

## 📊 Project Metrics

- **Lines of Code**: ~5,000+ lines of production Go code
- **Test Coverage**: 88%+ across handlers and business logic
- **Iterations Completed**: 25+ feature iterations
- **APIs Implemented**: 2 (REST + gRPC with 7-8 methods each)
- **Storage Backends**: 3 (memory, file, PostgreSQL)
- **Middleware Components**: 6+ (auth, logging, compression, content-type, etc.)
