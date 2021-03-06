package main

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"strconv"
	"time"
)

var msgQueue = make(chan string, 1000)
var known_ips = make(map[string]net.Conn)
var known_ips_lock = make(chan int, 1)
var uid = strconv.Itoa(rand.Int())
var prot = "udp"
var start_time time.Time = time.Now()
var started = false
var counter = 0
var primeSize int64 = 100

func sendContinuosly() {
	for {
		msg := <-msgQueue

		for _, conn := range known_ips {
			conn.Write([]byte(msg))
		}
	}
}

func w84c() {
	addr := net.UDPAddr{
		Port: 8080,
		IP:   net.ParseIP("0.0.0.0"),
	}
	pc, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Print("Error on listenning: ")
		fmt.Println(err)
		return
	}
	for {
		buf := make([]byte, 1024)
		n, _, err := pc.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		if buf[0] == 'M' {
			if !started {
				start_time = time.Now()
				started = true
				counter = 0
			}
			counter++
			if counter == 10000 {
				duration := start_time.Sub(time.Now())
				fmt.Println("Duration", duration)
			}
			if counter%100 == 0 {
				fmt.Println(counter, n)
			}
			//fmt.Println(string(buf[1:(n - 1)]))
		} else if buf[0] == 'I' {

		}
	}
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
	for i := 10000; i > 0; i-- {
		testNumber := getNumber(primeSize)
		if isPrime(testNumber) {
			return testNumber
		}
	}
	return big.NewInt(2)
}

func readToConnect() {
	known_ips["10.69.24.1"], _ = net.Dial("udp", "10.69.24.1:8080")
	known_ips["10.69.0.1"], _ = net.Dial("udp", "10.69.0.2:8080")
	for {
		msg := "MFrom " + uid + "\tPrime " + getPrime().String() + "#"
		msgQueue <- msg
	}
}

func ipSyncer() {
	for {
		<-known_ips_lock
		for key, _ := range known_ips {
			msgQueue <- "I" + key + "#"
		}
		known_ips_lock <- 1
		time.Sleep(5 * time.Second)
	}
}

func main() {
	known_ips_lock <- 1
	rand.Seed(time.Now().UnixNano())
	uid = strconv.Itoa(rand.Int())
	fmt.Println("Launch" + uid)
	go w84c()
	go readToConnect()
	go ipSyncer()
	go sendContinuosly()
	for {
		<-known_ips_lock
		for _, value := range known_ips {
			fmt.Println(value)
		}
		known_ips_lock <- 1
		time.Sleep(2 * time.Second)
	}
}
