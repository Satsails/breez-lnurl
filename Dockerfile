# Stage 1: Build the Go application and download the migration tool
FROM golang:1.24-alpine AS builder

# Install curl, which we need to download the migration tool
RUN apk add --no-cache curl

WORKDIR /app

# Download and install the migrate-cli tool. Using a specific version ensures reproducible builds.
# Check for the latest version at https://github.com/golang-migrate/migrate/releases
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz
RUN mv migrate /usr/local/bin/

# Copy go modules files and download dependencies to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application, creating a static binary named 'breez-lnurl'
RUN CGO_ENABLED=0 go build -o /app/breez-lnurl .

# ---

# Stage 2: Create the final, lightweight production image
FROM alpine:latest
\
# The migrate CLI with the postgres driver needs this library to run on Alpine Linux
RUN apk --no-cache add libc6-compat

WORKDIR /root/

# Copy the compiled application, migration tool, migration files, and startup script from the builder stage
COPY --from=builder /app/breez-lnurl .
COPY --from=builder /usr/local/bin/migrate .
COPY --from=builder /app/persist/migrations ./persist/migrations
COPY run.sh .
RUN chmod +x run.sh

# Set the command to run when the container starts
CMD ["./run.sh"]