FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/web-qa .

FROM mcr.microsoft.com/playwright:v1.57.0-noble

WORKDIR /app

COPY --from=builder /out/web-qa /app/web-qa
RUN chown -R pwuser:pwuser /app

USER pwuser

ENTRYPOINT ["/app/web-qa"]
