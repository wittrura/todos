# Stage 1: Build the Go application
FROM golang:1.25-alpine AS builder

# Creates an app directory to hold your app’s source code
WORKDIR /app
 
# Effectively tracks changes within go.mod and go.sum files
COPY go.mod go.sum ./
 
# Installs Go dependencies
RUN go mod download

# Copies your source code into the app directory
COPY . .
 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/api ./cmd/api
 
# Stage 2: Create a minimal production image
# this is an example of a multi-stage build
# we start with the big golang image, but slim down to something small just to run
# this finished binary

FROM alpine:latest

WORKDIR /root/

# Tells Docker which network port your container listens on
ENV PORT=8080
EXPOSE 8080

COPY --from=builder /bin/api /api

ENTRYPOINT ["/api"]