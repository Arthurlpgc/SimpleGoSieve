package main
import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	conn net.Conn
	status chan int
	msgQueue chan string
}

var known_ips = make(map[string]Client)
var known_ips_lock = make(chan int, 1)
var uid = strconv.Itoa(rand.Int())

func sendContinuosly(client Client) {
	for {
		msg := <- client.msgQueue
		client.conn.Write([]byte(msg))
	}
}

func readContinuosly(client Client) {
	reader := bufio.NewReader(client.conn)
	for {
		entry, err := reader.ReadBytes('#')
		str := string(entry)
		if err != nil {
			log.Fatal(err)
			return
		}
		str = strings.Trim(str, "#")
		if str[0] == 'M' {
			fmt.Println(str[1:])
		} else if str[0] == 'I' {
			addIP(str[1:])
		}
	}
}

func handleConnection(conn net.Conn, ip string) {
	ip = strings.Split(ip,":")[0]
	<- known_ips_lock 
	known_ips[ip] = Client{conn: conn, status: make(chan int), msgQueue: make(chan string, 1000)}
	known_ips_lock <- 1
	client := known_ips[ip]
	go sendContinuosly(client)
	go readContinuosly(client)
}

func w84c() {
    ln, err := net.Listen("tcp", ":8080")
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
	_, ok := known_ips[ip] 
	if ok {
		// ip already known
		return
	}
	if idCheck(ip) {
		return
	}
	conn, err := net.Dial("tcp", ip + ":8080")
	if err != nil {
		// handle error
		return
	}
	go handleConnection(conn, ip)
}

func readToConnect() {
	devices := 1
	ips := make([]string, devices)

	time.Sleep(5 * time.Second)
	for i := 0; i < devices; i++ {
		ips[i] = "10.69.24." + strconv.Itoa(i + 1)
	}
	for _, ip := range ips {
		addIP(ip)
		fmt.Println("Ip added " + ip)
	}
	fmt.Println("Ips added")
	for {
		<- known_ips_lock 
		msg := "Malo" + uid + "/" + strconv.Itoa(rand.Int()) + "#"
		for _, value := range known_ips {
			value.msgQueue <- msg
		}
		known_ips_lock <- 1
		time.Sleep(1 * time.Second)
	}
}

func ipSyncer() {
	for {
		<- known_ips_lock 
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
	conn, err := net.Dial("tcp", ip + ":8081")
	if err != nil {
		return true
	}
	id := make([]byte, 100) 
	conn.Read(id)
	return strings.Trim(string(id), string(id)[99:])== uid
}

func idChecker() {
	ln, _ := net.Listen("tcp", ":8081")
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
	for {}
}