package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

  conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
  defer conn.Close()

  connBuffer := make([]byte, 4096)
  _, err = conn.Read(connBuffer)
	if err != nil {
		fmt.Println("Error reading data from connection: ", err.Error())
		os.Exit(1)
	}

  response := "HTTP/1.1 200 OK\r\n\r\n"
  _, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing data to connection: ", err.Error())
		os.Exit(1)
	}
}
