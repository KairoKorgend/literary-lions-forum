FROM golang:1.21.6

WORKDIR /app/

COPY . .

WORKDIR /app/backend/cmd

RUN go build -o app .

EXPOSE 3000

CMD ["./app"]