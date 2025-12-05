FROM golang:1.23-alpine AS builder

WORKDIR /app

# добавить poppler
RUN apk update && apk add --no-cache poppler-utils

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# --- Runtime stage ---
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/main .
COPY migrations ./migrations

EXPOSE 8080

CMD ["./main"]
