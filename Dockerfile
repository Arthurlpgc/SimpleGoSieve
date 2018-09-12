FROM golang
RUN go get github.com/streadway/amqp
COPY . .
CMD go run rmq.go