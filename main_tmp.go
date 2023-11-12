package main

// import (
// 	"encoding/gob"
// 	"fmt"
// 	"log"
// 	"net"
// 	"time"
// )

// type P struct {
// 	M, N int64
// 	S_A  [2]int
// 	In   IN
// }

// type IN struct {
// 	V int
// }

// type NP struct {
// 	Req_Type int
// 	Req_From Address // ip and port number
// 	// Transaction      Transaction
// 	// Block            Block
// 	// Merkel_Tree      MerkelTree
// 	// Ip_Port_List     []Address
// 	// Merkel_Tree_Hash Hash
// 	// Block_Hash       Hash
// }

// type Add struct {
// 	Ip   uint32
// 	Port uint16
// }

// var port int = 8084

// func client_main() {
// 	fmt.Println("start client")
// 	conn, err := net.DialTCP("tcp", &net.TCPAddr{Port: 10000}, &net.TCPAddr{Port: port})
// 	if err != nil {
// 		log.Fatal("Connection error", err)
// 	}
// 	encoder := gob.NewEncoder(conn)
// 	// p := &P{M: 1, N: 2, S_A: [2]int{1, 2}, In: IN{1}}
// 	// p := &[]int{1, 2, 3, 4, 5, 6}
// 	p := &NP{Req_Type: 1}
// 	fmt.Println("Sending : ", p)
// 	encoder.Encode(p)
// 	conn.Close()
// 	fmt.Println("done")
// }

// func handleConnection(conn net.Conn) {
// 	dec := gob.NewDecoder(conn)
// 	// p := &P{}
// 	p := &NP{}
// 	// p := make([]int, 0, 1)
// 	dec.Decode(&p)
// 	fmt.Println("Received : ", p)
// 	fmt.Printf("%s", conn.RemoteAddr().String())
// 	conn.Close()
// }

// func server_main() {
// 	fmt.Println("start")
// 	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
// 	if err != nil {
// 		// handle error
// 	}
// 	for {
// 		conn, err := ln.Accept() // this blocks until connection or error
// 		if err != nil {
// 			// handle error
// 			continue
// 		}
// 		go handleConnection(conn) // a goroutine handles conn so that the loop can accept other connections
// 	}
// }

// func main() {
// 	go server_main()
// 	time.Sleep(time.Second * 5)
// 	go client_main()
// 	time.Sleep(time.Second * 5)
// }
