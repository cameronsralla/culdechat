FROM golang:1.24-alpine AS base

RUN apk add --no-cache git bash build-base

WORKDIR /app/api

# Cache deps using api module files
COPY api/go.mod api/go.sum ./
RUN go mod download

# Copy only the api directory contents (module root)
COPY api/ ./

# Default command runs the dev server; overridden by compose if needed
CMD ["go", "run", "./cmd/main.go"]


