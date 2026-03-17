FROM golang:1.23-alpine AS builder
WORKDIR /app
ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o product-service ./cmd

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/product-service /app/product-service
COPY --from=builder /app/config /app/config
EXPOSE 8083
CMD ["/app/product-service"]
