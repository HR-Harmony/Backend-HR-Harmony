FROM golang:1.21-alpine

WORKDIR /goapp

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

CMD ["./main"]