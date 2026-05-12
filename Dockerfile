# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/nomo-api ./cmd/api

FROM alpine:3.21
RUN adduser -D -H nomo
USER nomo
COPY --from=build /out/nomo-api /nomo-api
EXPOSE 8080
CMD ["/nomo-api"]
