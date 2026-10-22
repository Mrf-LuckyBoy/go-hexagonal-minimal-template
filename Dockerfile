# syntax=docker/dockerfile:1

# ---- build ----
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd

# ---- runtime ----
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=build /out/app /app/app
COPY --from=build /src/config /app/config
ENV APP_ENV=prod

# Uncomment once you add an HTTP server (match http.port in config)
# EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
