FROM golang:1.26-alpine AS build

WORKDIR /workspace
COPY . .

WORKDIR /workspace/src/go-microservices/payment-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/payment-service ./cmd/payment-service

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=build /bin/payment-service /usr/local/bin/payment-service

EXPOSE 8083
ENTRYPOINT ["/usr/local/bin/payment-service"]
