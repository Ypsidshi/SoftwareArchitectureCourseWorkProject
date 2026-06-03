FROM golang:1.26-alpine AS build

WORKDIR /workspace
COPY . .

WORKDIR /workspace/src/go-microservices/auth-service
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /bin/auth-service ./cmd/auth-service

FROM alpine:3.22

RUN apk add --no-cache wget \
    && addgroup -S app && adduser -S app -G app
USER app

COPY --from=build /bin/auth-service /usr/local/bin/auth-service

EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/auth-service"]
