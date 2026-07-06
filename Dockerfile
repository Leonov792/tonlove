FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY backend/go.mod ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o server .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
COPY frontend/ ./frontend/
EXPOSE 8080
CMD ["./server"]
