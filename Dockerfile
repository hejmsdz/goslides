FROM golang:1.22 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -v -o server

FROM alpine
RUN apk add --no-cache libc6-compat
WORKDIR /app
ENV PORT=8000
ENV GIN_MODE=release
COPY --from=builder /app/server .
COPY --from=builder /app/fonts fonts
COPY --from=builder /app/public public
EXPOSE 8000
CMD ["/app/server"]
