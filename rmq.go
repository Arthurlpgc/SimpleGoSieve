package main

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"time"
	"math"
	"github.com/streadway/amqp"
)

var uid = "0"
var start_time time.Time = time.Now()
var started = false
var counter = 0
var primeSize int64 = 100
var size int = 0
var times [100]float64

func w84c() {
	conn, _ := amqp.Dial("amqp://rabbitmq:5672/")
	defer conn.Close()
	ch, _ := conn.Channel()
	defer ch.Close()
	q, _ := ch.QueueDeclare(
		os.Getenv("receivequeue"), // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	msgs, _ := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	  )
	for msg := range msgs {
		var buf = msg.Body
		if buf[0] == 'M' {
			if !started {
				start_time = time.Now()
				started = true
				counter = 0
			}
			counter++
			if counter == 10000 {
				duration := start_time.Sub(time.Now())
				times[size] = (float64(duration / time.Millisecond))
				size++
				calculos(size)
				fmt.Println("Duration", duration)
				counter = 0
				started = false
			}
			//fmt.Println(string(buf[1:(2 - 1)]))
		} else if buf[0] == 'I' {

		}
	}
}

func calculos(size int) {
	if(size != 50) {
		return;
	}

	mean := 0.0;
	sd := 0.0;
	for i := 0; i < size; i++ {
		mean += times[i]
		
	}

	mean = mean/float64(size)
	
	for i := 0; i < size; i++ {
		sd += math.Pow(times[i] - mean, 2)
	}
	
	fmt.Println("MEAN", mean, "SD", math.Sqrt(sd/float64(size)))

}

func getNumber(expo int64) *big.Int {
	//Gets random big number smaller than 2 to expo
	ret := big.NewInt(0)
	bound := big.NewInt(2)
	expobig := big.NewInt(expo)
	bound.Exp(bound, expobig, ret)
	n, _ := cryptorand.Int(cryptorand.Reader, bound)
	return n
}

func isPrime(x *big.Int) bool {
	base := big.NewInt(2)
	expo := big.NewInt(0)
	one := big.NewInt(1)
	expo.SetBytes(x.Bytes())
	expo.Sub(expo, one)
	base.Exp(base, expo, x)
	return base.Cmp(one) == 0
}

func getPrime() *big.Int {
	return getNumber(primeSize)
}

func readToConnect() {
	conn, err := amqp.Dial("amqp://rabbitmq:5672/")
	fmt.Println(err)
	defer conn.Close()
	ch, _ := conn.Channel()
	defer ch.Close()
	q, _ := ch.QueueDeclare(
		os.Getenv("sendqueue"), // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	for {
		msg := "MFrom " + uid + "\tPrime " + getPrime().String() + "#"
		_ = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing {
			  ContentType: "text/plain",
			  Body:        []byte(msg),
			})
	}
}

func main() {

	time.Sleep(10 * time.Second)
	rand.Seed(time.Now().UnixNano())
	uid = strconv.Itoa(rand.Int())
	fmt.Println("Launch" + uid)
	go w84c()
	go readToConnect()
	for {
		time.Sleep(2 * time.Second)
	}
}
