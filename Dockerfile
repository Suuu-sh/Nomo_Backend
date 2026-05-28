# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/tomo-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/tomo-notification-worker ./cmd/notification_worker

FROM alpine:3.21
RUN adduser -D -H tomo
USER tomo
COPY --from=build /out/tomo-api /tomo-api
COPY --from=build /out/tomo-notification-worker /tomo-notification-worker
EXPOSE 8080
CMD ["/tomo-api"]
