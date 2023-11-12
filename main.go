package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func main() {

	bootstrap_address := Address{Port: 8080}
	go peer_main(bootstrap_address, 20, true, false, true, Address{}, 4, 6)
	for i := 8081; i <= 8083; i++ {
		go peer_main(Address{Port: uint16(i)}, 20, false, i%2 == 1, i%2 == 0, bootstrap_address, 4, 6)
	}

	time.Sleep(30 * time.Second)

	go peer_main(Address{Port: 8086}, 20, false, true, true, bootstrap_address, 4, 6)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter text: ")
	reader.ReadString('\n')
}
