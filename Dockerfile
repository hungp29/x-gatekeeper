# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

# git required for go mod download when fetching modules from VCS
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /xgatekeeper ./cmd/xgatekeeper

# Runtime stage
FROM gcr.io/distroless/base-debian12
COPY --from=builder /xgatekeeper /xgatekeeper
EXPOSE 8080
ENV HTTP_PORT=8080
USER 65534:65534
ENTRYPOINT ["/xgatekeeper"]
