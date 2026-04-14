# Stage 1: Build the Svelte frontend
FROM node:20-bookworm AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Stage 2: Build the Go project
FROM golang:1.25-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy the built frontend from stage 1
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
# Build the main binary (without Wails desktop support, which requires CGO and GTK)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/web-qa .

# Stage 3: Final image
FROM mcr.microsoft.com/playwright:v1.57.0-noble
WORKDIR /app
COPY --from=builder /out/web-qa /app/web-qa
RUN chown -R pwuser:pwuser /app
USER pwuser
ENTRYPOINT ["/app/web-qa"]
