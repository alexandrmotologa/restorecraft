# Multi-stage build for restorecraft
# Stage 1: Build React 19 Frontend
FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go Binary
FROM golang:1.23-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/restorecraft ./cmd/restorecraft

# Stage 3: Runtime Container
FROM alpine:3.20
RUN apk add --no-cache ca-certificates docker-cli tzdata
WORKDIR /app
COPY --from=go-builder /app/bin/restorecraft /usr/local/bin/restorecraft
EXPOSE 8080
VOLUME ["/app/data", "/app/backups"]
ENTRYPOINT ["restorecraft"]
CMD ["serve", "--port", "8080", "--host", "0.0.0.0"]
