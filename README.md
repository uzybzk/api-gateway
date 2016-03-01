# API Gateway

A simple HTTP reverse proxy and API gateway built with Go.

## Features

- Request routing and proxying
- CORS support
- Request logging
- Health checks
- Environment-based configuration
- Load balancing ready

## Usage

```bash
go run main.go
```

## Configuration

Environment variables:
- `PORT` - Gateway port (default: 8080)
- `USERS_SERVICE` - Users service URL
- `POSTS_SERVICE` - Posts service URL  
- `AUTH_SERVICE` - Auth service URL

## Routes

- `/api/users/*` → Users microservice
- `/api/posts/*` → Posts microservice
- `/api/auth/*` → Authentication microservice
- `/health` → Gateway health check

## Docker

```dockerfile
FROM golang:1.7-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o gateway ./main.go

FROM alpine:3.4
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/gateway .
EXPOSE 8080
CMD ["./gateway"]
```

## Docker Compose Example

```yaml
version: '3'
services:
  gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      - USERS_SERVICE=http://users:8080
      - POSTS_SERVICE=http://posts:8080
      - AUTH_SERVICE=http://auth:8080
    depends_on:
      - users
      - posts
      - auth
```

## Features TODO

- [ ] Rate limiting
- [ ] Circuit breaker
- [ ] Request/response transformation
- [ ] Authentication middleware
- [ ] Metrics collection
- [ ] Service discovery integration