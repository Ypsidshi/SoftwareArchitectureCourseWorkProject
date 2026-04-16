FROM golang:1.26-alpine AS build

WORKDIR /workspace
COPY . .

WORKDIR /workspace/src/go-microservices/deal-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/deal-service ./cmd/deal-service

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=build /bin/deal-service /usr/local/bin/deal-service

EXPOSE 8082
ENTRYPOINT ["/usr/local/bin/deal-service"]
