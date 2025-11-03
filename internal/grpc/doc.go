// Package server provides gRPC server implementation for the URL shortener service.
//
// # Overview
//
// This package implements a gRPC API that mirrors the HTTP REST API functionality,
// providing efficient binary protocol communication with features like:
//   - URL shortening (single and batch operations)
//   - URL retrieval and redirection
//   - User-specific URL management
//   - Service statistics
//   - Health checks
//
// # Authentication
//
// The gRPC server uses JWT-based authentication similar to the HTTP API:
//   - JWT tokens are passed via gRPC metadata with key "jwt"
//   - If no token is provided, a new one is generated automatically
//   - User ID is extracted from the token and stored in context
//   - Public methods (Ping, GetShortURL, GetStats) bypass authentication
//
// # Public vs Private Methods
//
// Public methods (no authentication required):
//   - Ping: Health check endpoint
//   - GetShortURL: Retrieve original URL by short URL (redirect)
//   - GetStats: Service statistics (requires trusted IP)
//
// Private methods (authentication required):
//   - CreateShortURL: Create a new short URL
//   - BatchCreateShortURL: Create multiple short URLs in one request
//   - ListShortURLs: Get all URLs for the authenticated user
//   - DeleteURLs: Soft-delete user's URLs
//
// # Error Handling
//
// The server uses gRPC status codes for error responses:
//   - codes.InvalidArgument: Invalid URL or request parameters
//   - codes.Unauthenticated: Missing or invalid JWT token
//   - codes.NotFound: URL not found or deleted
//   - codes.AlreadyExists: URL already exists (conflict)
//   - codes.PermissionDenied: Access denied (e.g., IP not in trusted subnet)
//   - codes.Internal: Server internal errors
//
// # Configuration
//
// The server requires:
//   - GRPCAddr: gRPC server listen address (default: :9090)
//   - BaseURL: Base URL for shortened links (e.g., http://localhost:8080)
//   - SecretKey: JWT signing key (required, must be set via environment)
//   - TrustedSubnet: CIDR for trusted IPs (optional, for GetStats)
//
// # Example Usage
//
//	// Create gRPC server
//	cfg := &config.Config{
//	    GRPCAddr: ":9090",
//	    BaseURL:  "http://localhost:8080",
//	    SecretKey: os.Getenv("SECRET_KEY"),
//	}
//
//	repo := repository.NewRepository(cfg)
//	urlChecker := validate.NewService()
//	jwtService := jwt.NewService(cfg.SecretKey, 24*time.Hour)
//
//	server := NewShortenerServer(cfg, repo, urlChecker, jwtService)
//
//	// Run server
//	go func() {
//	    if err := server.Run(); err != nil {
//	        log.Fatal(err)
//	    }
//	}()
//
//	// Graceful shutdown
//	<-ctx.Done()
//	server.Shutdown()
//
// # Proto Definition
//
// The gRPC API is defined in api/grpc/proto/shortener/v1/shortener.proto
// Generated code is in api/grpc/gen/shortener/v1/
//
// To regenerate gRPC code:
//
//	cd api/grpc && buf generate
//
// # Middleware
//
// The server uses the following interceptors:
//   - UnaryAuthInterceptor: JWT authentication and user ID injection
//
// Auth interceptor flow:
//  1. Check if method is public (skip auth if true)
//  2. Extract JWT from metadata
//  3. If no token, generate new one and add to outgoing metadata
//  4. If token exists, validate and extract user ID
//  5. Add user ID to context for handler access
//
// # Thread Safety
//
// The ShortenerServer is safe for concurrent use. All handlers are stateless
// and thread-safe, delegating state management to the repository layer.
//
// # Logging
//
// All operations are logged using structured logging (zap):
//   - Debug: Authentication events, method calls
//   - Info: Successful operations, user actions
//   - Warn: Configuration issues, empty requests
//   - Error: Failures, invalid tokens, repository errors
package server
