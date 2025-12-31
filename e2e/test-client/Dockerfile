# Build from project root to access rgrpc and proto packages
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install protoc for proto generation if needed
RUN apk add --no-cache protobuf protoc make

# Copy entire project
COPY . .

# Generate go.sum files for proto package (standalone module)
WORKDIR /build/e2e/proto
RUN go mod tidy

# Generate go.sum for root module (contains rgrpc package)
WORKDIR /build
RUN go mod tidy

# Build test-client (will use replace directive to find rgrpc in root module)
WORKDIR /build/e2e/test-client
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o test-client ./main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/e2e/test-client/test-client .

EXPOSE 8080

CMD ["./test-client"]

