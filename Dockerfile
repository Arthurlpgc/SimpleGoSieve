FROM golang
COPY . .
CMD go run tcp.go