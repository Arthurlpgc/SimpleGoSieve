FROM golang
RUN go get github.com/streadway/amqp
COPY . .
ENV limit=100000
CMD go run rmq.go