FROM golang:1.26-alpine AS build

WORKDIR /workspace
COPY . .

WORKDIR /workspace/src/go-microservices/deal-service
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /bin/deal-service ./cmd/deal-service

FROM alpine:3.22

RUN apk add --no-cache wget \
    && addgroup -S app && adduser -S app -G app
USER app

COPY --from=build /bin/deal-service /usr/local/bin/deal-service

EXPOSE 8082
ENTRYPOINT ["/usr/local/bin/deal-service"]
