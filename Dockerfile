FROM golang:1.22.4-alpine3.20

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY src ./src

RUN CGO_ENABLED=0 GOOS=linux go build -C src -o /pixel-tactics-match

EXPOSE 8000
CMD ["/pixel-tactics-match"]
