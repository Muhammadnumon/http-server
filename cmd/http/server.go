package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

const (
	httpStatusOK       = "200 OK"
	httpStatusNotFound = "404 not found"
)

func main() {
	file, err := os.OpenFile("server.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("can't create log file, %v", err)
	}
	defer func() {
		log.Print("closing file")
		err := file.Close()
		if err != nil {
			log.Printf("can't close file %v", err)
			return
		}
		log.Print("file closed")
	}()

	log.SetOutput(file)

	host := "0.0.0.0"
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "9999"
	}

	log.Print("start listening")
	listener, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		log.Printf("can't listen %v", err)
		return
	}
	log.Print("listening started")
	for {
		log.Print("start accepting")
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("can't accept connection: %v", err)
			return
		}
		log.Print("connection accepted")
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	log.Print("start reading")
	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("can't read string: %v", err)
		return
	}
	log.Print("request: ", requestLine)

	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(parts) != 3 {
		return
	}

	method, request, protocol := parts[0], parts[1], parts[2]

	if method != "GET" || protocol != "HTTP/1.1" {

	}

	requestTypes := []string{"/", "/image.jpg", "/image.png", "/page.html", "/file.txt", "/sample.pdf"}
	contentNames := []string{"index.html", "photo.jpg", "image.png", "another-index.html", "text-file.txt", "go.pdf"}
	contentTypes := []string{"text/html", "image/jpg", "image/png", "text/html", "text/plain", "application/pdf"}

	var contentType string
	if strings.Contains(request, "?download") {
		contentType = "application/octet-stream"
		request = strings.Replace(request, "?download", "", -1)
	}
	for indx, requestType := range requestTypes {
		if request != requestType {
			continue
		}
		if contentType == "application/octet-stream" {
			sendContent(method, request, protocol, contentNames[indx], contentType, httpStatusOK, conn)
			return
		}
		sendContent(method, request, protocol, contentNames[indx], contentTypes[indx], httpStatusOK, conn)
		return
	}
	sendContent(method, request, protocol, "404-index.html", "text/html", httpStatusNotFound, conn)
}

func sendContent(method, request, protocol, contentName, contentType, status string, conn net.Conn) {
	log.Printf("method: %s; request: %s; protocol: %s", method, request, protocol)
	pageContent, err := ioutil.ReadFile("cmd/http/files/" + contentName)
	if err != nil {
		log.Printf("can't read from file: %s, %v", contentName, err)
		return
	}
	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString(fmt.Sprintf("HTTP/1.1 %s\r\n", status))
	if err != nil {
		log.Printf("can't write %v", err)
		return
	}
	_, err = writer.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(pageContent)))
	if err != nil {
		log.Printf("can't write %v", err)
		return
	}
	_, err = writer.WriteString(fmt.Sprintf("Content-Type: %s\r\n", contentType))
	log.Print(contentType)
	if err != nil {
		log.Printf("can't write %v", err)
		return
	}
	_, err = writer.WriteString("Connection: close\r\n")
	if err != nil {
		log.Printf("can't write %v", err)
		return
	}
	_, err = writer.WriteString("\r\n")
	if err != nil {
		log.Printf("can't write %v", err)
		return
	}
	_, err = writer.Write(pageContent)
	if err != nil {
		log.Printf("can't write %v", err)
		return
	}
	err = writer.Flush()
	if err != nil {
		log.Printf("can't flush %v", err)
	}
}
