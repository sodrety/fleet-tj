FROM golang:1.24.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /fleet-backend ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /fleet-backend /fleet-backend
COPY config/geofence.yaml /config/geofence.yaml
ENTRYPOINT ["/fleet-backend"]
