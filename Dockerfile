FROM golang:1.23.4

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o RedicaMq

CMD ["./RedicaMq"]

