package main

import (
	"bufio"
	cryptorand "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	conn     net.Conn
	status   chan int
	msgQueue chan string
}

var known_ips = make(map[string]Client)
var known_ips_lock = make(chan int, 1)
var uid = strconv.Itoa(rand.Int())
var prot = "tcp"
var start_time time.Time = time.Now()
var started = false
var counter = 0
var primeSize int64 = 100

func sendContinuosly(client Client) {
	for {
		msg := <-client.msgQueue
		_, err := client.conn.Write([]byte(msg))
		if err != nil {
			fmt.Println("sendContinuosly failed: ", err)
		}
	}
}

func readContinuosly(client Client) {
	reader := bufio.NewReader(client.conn)
	for {
		entry, err := reader.ReadString('#')
		str := string(entry)
		if err != nil {
			log.Fatal(err)
			return
		}
		str = strings.Trim(str, "#")
		if str[0] == 'M' {
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
				fmt.Println(counter)
			}
			//fmt.Println(str[1:])
		} else if str[0] == 'I' {
			addIP(str[1:])
		}
	}
}

func handleConnection(conn net.Conn, ip string) {
	ip = strings.Split(ip, ":")[0]
	<-known_ips_lock
	known_ips[ip] = Client{conn: conn, status: make(chan int), msgQueue: make(chan string, 1000)}
	known_ips_lock <- 1
	client := known_ips[ip]
	go sendContinuosly(client)
	go readContinuosly(client)
}

func w84c() {
	ln, err := net.Listen(prot, ":8080")
	if err != nil {
		fmt.Print("Error on listenning: ")
		fmt.Println(err)
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		ip := conn.RemoteAddr().String()
		go handleConnection(conn, ip)
	}
}

func addIP(ip string) {
	if idCheck(ip) {
		fmt.Println(ip + " Self")
		return
	}
	_, ok := known_ips[ip]
	if ok {
		fmt.Println(ip + " Known")
		return
	}
	conn, err := net.Dial(prot, ip+":8080")
	if err != nil {
		fmt.Println(ip+" Conn Error", err)
		return
	}
	go handleConnection(conn, ip)
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
	addIP("10.69.24.1")
	fmt.Println("Ips added")
	for {
		msg := "MFrom " + uid + "\tPrime " + getPrime().String() + "#"
		<-known_ips_lock
		for _, value := range known_ips {
			value.msgQueue <- msg
		}
		known_ips_lock <- 1
		time.Sleep(1 * time.Nanosecond)
	}
}

func ipSyncer() {
	for {
		<-known_ips_lock
		for _, value := range known_ips {
			for key, _ := range known_ips {
				value.msgQueue <- "I" + key + "#"
			}
		}
		known_ips_lock <- 1
		time.Sleep(5 * time.Second)
	}
}

func idCheck(ip string) bool {
	conn, err := net.Dial(prot, ip+":8081")
	for i := 0; i < 10 && err != nil; i++ {
		conn, err = net.Dial(prot, ip+":8081")
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return true
	}
	id := make([]byte, 100)
	conn.Read(id)
	return strings.Trim(string(id), string(id)[99:]) == uid
}

func idChecker() {
	ln, _ := net.Listen(prot, ":8081")
	for {
		conn, _ := ln.Accept()
		conn.Write([]byte(uid))
		conn.Close()
	}
}

func main() {
	known_ips_lock <- 1
	rand.Seed(time.Now().UnixNano())
	uid = strconv.Itoa(rand.Int())
	fmt.Println("Launch" + uid)
	go idChecker()
	go w84c()
	go readToConnect()
	go ipSyncer()
	for {
		<-known_ips_lock
		for _, value := range known_ips {
			fmt.Println(value.conn)
		}
		known_ips_lock <- 1
		time.Sleep(2 * time.Second)
	}
}
