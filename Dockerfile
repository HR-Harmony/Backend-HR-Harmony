FROM golang:1.21.0-alpine

WORKDIR /app


COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o main.app .

EXPOSE 8080
ENV PORT 8080
ENV HOSTNAME "0.0.0.0"

CMD ["/app/main.app"]