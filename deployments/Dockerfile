# Stage 1: Build frontend (placeholder for Phase 2)
# FROM node:20-alpine AS frontend
# WORKDIR /app/web
# COPY web/package.json web/package-lock.json ./
# RUN npm ci
# COPY web/ ./
# RUN npm run build

# Stage 2: Build backend
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/polaris ./cmd/polaris

# Stage 3: Runtime
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/polaris .
COPY configs/ ./configs/
# COPY --from=frontend /app/web/dist ./web/dist
EXPOSE 8080
ENTRYPOINT ["/app/polaris"]
CMD ["serve", "--config", "/app/configs/polaris.yaml"]
