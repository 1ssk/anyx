FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o invite ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/invite ./invite
COPY web ./web
COPY .env ./.env
EXPOSE 8080
CMD ["./invite"]
