# Build from e2e directory to access proto package
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install protoc for proto generation if needed
RUN apk add --no-cache protobuf protoc make

# Copy entire e2e directory
COPY e2e/ ./e2e/

# Generate go.sum files for proto package
WORKDIR /build/e2e/proto
RUN go mod tidy

# Build test-server
WORKDIR /build/e2e/test-server
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o test-server ./main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/e2e/test-server/test-server .

EXPOSE 50051

CMD ["./test-server"]

