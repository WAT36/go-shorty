# Build stage
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o shorty ./cmd/shorty

# Run stage
FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/shorty /usr/local/bin/shorty
COPY web ./web
# データはボリューム化
VOLUME ["/app/data"]
ENV PORT=8080
EXPOSE 8080
CMD ["shorty"]
