package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type req struct {
  method string
  path string
  version string
  headers map[string]string
}

var dirFlag = flag.String("directory", ".", "directory to serve files from")

func main() {
  flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
    os.Exit(1)
	}

  for {
    conn, err := l.Accept()
    if err != nil {
      fmt.Println("Error accepting connection: ", err.Error())
      continue
    }
    go handleConnection(conn)
  }
}

func handleConnection(conn net.Conn) {
  defer conn.Close()

  connReader := bufio.NewReader(conn)

  reqInfo, err := connReader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading data from connection: ", err.Error())
		sendInternalServerErrorResponse(conn)
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

  switch req.method {
  case "GET":
    handleGET(conn, req)
  case "POST":
    handlePOST(conn, connReader, req)
  default:
    panic("That method hasn't been implemented!")
  }
}

func handleGET(conn net.Conn, req req) {
  pathParts := parsePathParts(req)

  if req.path == "/" {
    sendSuccessResponse(nil, "", conn)
  } else if req.path == "/user-agent" {
    sendSuccessResponse([]byte(req.headers["User-Agent"]), "text/plain", conn)
  } else if len(pathParts) == 2 && pathParts[0] == "files" {
    filePath := filepath.Join(*dirFlag, pathParts[1])
    _, err := os.Stat(filePath)
    if err != nil {
      fmt.Println("Error accessing requested file: ", err.Error())
      sendNotFoundResponse(conn)
      return
    }
    contents, err := os.ReadFile(filePath)
    if err != nil {
      fmt.Println("Error reading requested file: ", err.Error())
      sendInternalServerErrorResponse(conn)
    }
    sendSuccessResponse(contents, "application/octet-stream", conn)
  } else if len(pathParts) == 2 && pathParts[0] == "echo" {
    sendSuccessResponse([]byte(pathParts[1]), "text/plain", conn)
  } else {
    sendNotFoundResponse(conn)
  }
}

func handlePOST(conn net.Conn, connReader *bufio.Reader, req req) {
  pathParts := parsePathParts(req)
  if len(pathParts) == 2 && pathParts[0] == "files" {
    filePath := filepath.Join(*dirFlag, pathParts[1])
    file, err := os.Create(filePath)
    if err != nil {
      fmt.Println("Error creating file: ", err.Error())
      sendInternalServerErrorResponse(conn)
    }

    contentLength, err := strconv.Atoi(req.headers["Content-Length"])
    if err != nil {
      fmt.Println("Invalid Content-Length")
      sendInternalServerErrorResponse(conn)
      return
    }

    // write file 
    buffer := make([]byte, 4096)
    totalBytesRead := 0

    for totalBytesRead < contentLength {
      bytesRead, err := connReader.Read(buffer)
      if err != nil {
        if err != io.EOF {
          fmt.Println("Error reading posted file: ", err)
          sendInternalServerErrorResponse(conn)
        }
        break
      }
      totalBytesRead += bytesRead
      _, err = file.Write(buffer[:bytesRead])
      if err != nil {
        fmt.Println("Error writing file: ", err)
        sendInternalServerErrorResponse(conn)
        break
      }
    }
    sendCreatedResponse(conn)
  } else {
    sendNotFoundResponse(conn)
  }
}

func parsePathParts(req req) []string {
  parsedURL, err := url.Parse(req.path)
  if err != nil {
    fmt.Printf("Error parsing URL: %s\n", err)
  }
  pathParts := strings.SplitN(parsedURL.Path, "/", 3)
  if len(pathParts) > 0 && pathParts[0] == "" {
    pathParts = pathParts[1:]
  }
  return pathParts
}

func sendInternalServerErrorResponse(conn net.Conn) {
  response := []byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")
  _, err := conn.Write(response)
  if err != nil {
    fmt.Println("Error writing data to connection: ", err.Error())
    os.Exit(1)
  }
}

func sendNotFoundResponse(conn net.Conn) {
  response := []byte("HTTP/1.1 404 Not Found\r\n\r\n")
  _, err := conn.Write(response)
  if err != nil {
    fmt.Println("Error writing data to connection: ", err.Error())
    sendInternalServerErrorResponse(conn)
  }
}

func sendCreatedResponse(conn net.Conn) {
  response := []byte("HTTP/1.1 201 Created\r\n\r\n")
  _, err := conn.Write(response)
  if err != nil {
    fmt.Println("Error writing data to connection: ", err.Error())
    sendInternalServerErrorResponse(conn)
  }
}

func sendSuccessResponse(content []byte, contentType string, conn net.Conn) {
  response := []byte("HTTP/1.1 200 OK\r\n")
  if content != nil && contentType != "" {
    response = append(response, []byte(fmt.Sprintf("Content-Type: %s\r\n", contentType))...)
    response = append(response, []byte(fmt.Sprintf("Content-Length: %d\r\n", len(content)))...)
    response = append(response, []byte("\r\n")...)
    response = append(response, content...)
  }
  response = append(response, []byte("\r\n")...)
  _, err := conn.Write(response)
  if err != nil {
    fmt.Println("Error writing data to connection: ", err.Error())
    sendInternalServerErrorResponse(conn)
  }
}
