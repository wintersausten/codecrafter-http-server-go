package main

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
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

  parsedURL, err := url.Parse(req.path)
  if err != nil {
    fmt.Printf("Error parsing URL: %s\n", err)
  }
  pathParts := strings.SplitN(parsedURL.Path, "/", 3)
  if len(pathParts) > 0 && pathParts[0] == "" {
      pathParts = pathParts[1:]
  }

  if req.path == "/" {
    response := "HTTP/1.1 200 OK\r\n\r\n"
    _, err = conn.Write([]byte(response))
    if err != nil {
      fmt.Println("Error writing data to connection: ", err.Error())
      os.Exit(1)
    }
  } else if len(pathParts) == 2 && pathParts[0] == "echo" {
    response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(pathParts[1]), pathParts[1])
    _, err = conn.Write([]byte(response))
    if err != nil {
      fmt.Println("Error writing data to connection: ", err.Error())
      os.Exit(1)
    }
  } else {
    response := "HTTP/1.1 404 Not Found\r\n\r\n"
    _, err = conn.Write([]byte(response))
    if err != nil {
      fmt.Println("Error writing data to connection: ", err.Error())
      os.Exit(1)
    }
  }
}

