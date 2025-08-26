# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
# If go.sum existed, we'd copy it too
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/bitcoin-prices ./

# Runtime stage
FROM alpine:3.20
RUN adduser -D -H appuser
USER appuser
WORKDIR /app
COPY --from=build /out/bitcoin-prices /app/bitcoin-prices
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/app/bitcoin-prices"]

