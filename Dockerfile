FROM golang:1.21.2

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN go build -o main

CMD ["./main"]