FROM golang:1.21-alpine AS builder

WORKDIR /app

# Set env vars
ARG GOOS=linux
ARG GOARCH=amd64

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=${GOOS} \
    GOARCH=${GOARCH}

# Copy go.mod and go.sum and download dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code
COPY . .

# Build the application
RUN go build -o main .

# Final image to run the application
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Run the binary
CMD ["./main"]

