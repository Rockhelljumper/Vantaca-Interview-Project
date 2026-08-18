FROM golang:1.25-alpine AS build

WORKDIR /src
COPY api/go.mod api/go.sum* ./api/
WORKDIR /src/api
RUN go mod download

WORKDIR /src
COPY api ./api
COPY database ./database
WORKDIR /src/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/vantaca-api ./cmd/api

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app
COPY --from=build /out/vantaca-api /app/vantaca-api
COPY --from=build /src/database /app/database

USER app
EXPOSE 8080
ENTRYPOINT ["/app/vantaca-api"]
