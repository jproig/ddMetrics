FROM golang:latest

RUN mkdir /app
WORKDIR /app
COPY . .
RUN go build -o main .

CMD ["/app/main", "-loop_count=100", "-write_on_dir=/mnt"]