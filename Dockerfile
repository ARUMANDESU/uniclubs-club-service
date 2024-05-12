FROM golang:1.21 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o ./build/main ./cmd/

ENV ENV="dev"
ENV DATABASE_DSN="postgres://postgres:password@postgres:5432/clubdb"

ENV GRPC_PORT=44045
ENV GRPC_TIMEOUT=1h

ENV RABBITMQ_USER="admin"
ENV RABBITMQ_PASSWORD="admin"
ENV RABBITMQ_HOST="localhost"
ENV RABBITMQ_PORT="5672"


EXPOSE 44045

ENTRYPOINT ["./build/main"]