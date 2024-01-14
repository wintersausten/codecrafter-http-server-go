package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type req struct {
  method string
  path string
  version string
  headers map[string]string
}

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

  connReader := bufio.NewReader(conn)

  reqInfo, err := connReader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading data from connection: ", err.Error())
		os.Exit(1)
	}

  reqInfoParts := strings.Fields(reqInfo)

  req := req{}
  req.method = reqInfoParts[0]
  req.path = reqInfoParts[1]
  req.version = reqInfoParts[2]

  req.headers = map[string]string{}
  for {
    headerLine, err := connReader.ReadString('\n')
    if err != nil || headerLine == "\r\n" {
      break 
    }

    headerParts := strings.SplitN(headerLine, ": ", 2)
    if len(headerParts) == 2 {
      key := headerParts[0]
      value := headerParts[1]
      req.headers[key] = strings.TrimSpace(value)
    }
  }

  switch req.path {
  case "/":
    response := "HTTP/1.1 200 OK\r\n\r\n"
    _, err = conn.Write([]byte(response))
    if err != nil {
      fmt.Println("Error writing data to connection: ", err.Error())
      os.Exit(1)
    }
  default:
    response := "HTTP/1.1 404 Not Found\r\n\r\n"
    _, err = conn.Write([]byte(response))
    if err != nil {
      fmt.Println("Error writing data to connection: ", err.Error())
      os.Exit(1)
    }
  }
}
