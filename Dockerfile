FROM golang:1.21.0-alpine

WORKDIR /app


COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o main.app .

EXPOSE 8080

CMD ["/app/main.app"]