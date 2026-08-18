FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN go test ./... && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/northwind-mock ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/northwind-mock /northwind-mock

EXPOSE 8081
USER nonroot:nonroot
HEALTHCHECK --interval=5s --timeout=3s --start-period=3s --retries=5 CMD ["/northwind-mock", "healthcheck"]
ENTRYPOINT ["/northwind-mock"]
