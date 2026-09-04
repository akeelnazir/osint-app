# syntax=docker/dockerfile:1
FROM golang:1.27-alpine AS build
WORKDIR /src

# Cache deps.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build server binary.
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server
# Build seed binary.
RUN CGO_ENABLED=0 go build -o /out/seed ./cmd/seed

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /out/seed /app/seed
RUN mkdir -p /data/uploads
EXPOSE 8080
CMD ["/app/server"]
