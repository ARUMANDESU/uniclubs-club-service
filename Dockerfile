FROM golang:1.21 as builder

WORKDIR /app

# Copy the go.mod and go.sum files first and download the dependencies.
# This is done separately from copying the entire source code to leverage Docker cache
# and avoid re-downloading dependencies if they haven't changed.
COPY go.mod go.sum ./
RUN go mod download
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
RUN apt-get update && apt-get install -y libvips-dev

# Copy the rest of the application's source code.
COPY . .

# Build the application. This assumes you have a main package at the root of your project.
# Adjust the path to the main package if it's located elsewhere.
RUN CGO_ENABLED=1 GOOS=linux go build -o ./build/main ./cmd/

# Define environment variables for PostgreSQL and Redis connections.
# These values can be overridden when running the container.
ENV ENV="dev"
ENV DATABASE_DSN="postgres://postgres:password@postgres:5432/clubdb"
ENV GRPC_PORT=44045
ENV GRPC_TIMEOUT=1h
ENV RABBITMQ_USER="dsadsi21neoU@N!D"
ENV RABBITMQ_PASSWORD="Y98213KQSNDKJASKDLJNka"
ENV RABBITMQ_HOST="localhost"
ENV RABBITMQ_PORT="5672"
ENV AWS_REGION="us-east-1"
ENV AWS_ACCESS_KEY_ID="access_key_id"
ENV AWS_SECRET_ACCESS_KEY="secret_access_key"
ENV AWS_S3_BUCKET="bucket_name"

# Expose the port your application listens on.
EXPOSE 44045

# Run the application.
ENTRYPOINT ["./build/main"]
CMD ["migrate", "-path", "./migrations", "-database", "$DATABASE_DSN?sslmode=disable", "up"]