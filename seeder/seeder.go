package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
)

var peerList []string
var fileRequests = make(chan string)

func main() {
	// Listen for incoming connections
	// 10.5.226.133 under aau staff
	ln, err := net.Listen("tcp", "192.168.43.37:1234")
	if err != nil {
		fmt.Println("Error starting server: ", err)
		return
	}
	fmt.Println("Server started", ln.Addr())
	defer ln.Close()

	// go downloadTorrent()

	// Start goroutines to handle file requests and the web server
	go handleFileRequests()
	go handleWebServer()
	// go handleFileReceive()

	for {
		// Accept incoming connections
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())

		go handleConnection(conn)
	}
}

func handleFileRequests() {
	fmt.Println("Waiting for file requests...")
	for {
		fileName := <-fileRequests
		go downloadFile(fileName)
	}
}

func downloadFile(fileName string) {

	// get fileName extension
	fileExtension := filepath.Ext(fileName)

	if fileExtension == ".torrent" {
		fmt.Println("file with", fileExtension, "Torrent file detected")
		// go downloadTorrent(fileName)
	} else {
		fmt.Println("file with", fileExtension, "Regular file detected")
		// downloadRegularFile(fileName)
	}

	for _, peer := range peerList {
		conn, err := net.Dial("tcp", "192.168.43.117:8080")
		if err != nil {
			fmt.Println("Error connecting to peer: ", err)
			return
		}
		defer conn.Close()
		fmt.Println("peer connected", peer)

		fileContent, err := ioutil.ReadFile(fileName)

		// fmt.Println("file content", string(fileContent))

		if err != nil {
			fmt.Println("Error reading file:", err.Error())
			return
		}

		_, err = fmt.Fprintln(conn, fileName)
		if err != nil {
			fmt.Println("Error sending file name:", err.Error())
			return
		}

		fmt.Println("Sending Filename", string(fileName))

		_, err = conn.Write(fileContent)
		if err != nil {
			fmt.Println("Error sending file content:", err.Error())
			return
		}

		// fmt.Println("Sending File Content", string(fileContent))
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	// Get peer address
	peerAddr := conn.RemoteAddr().String()
	fmt.Println("Peer connected:", peerAddr)
	// Add peer to peer list
	peerList = append(peerList, peerAddr)
	fmt.Println("Peer connected:", peerAddr)
	defer func() {
		// Remove peer from peer list
		for i, p := range peerList {
			if p == conn.RemoteAddr().String() {
				peerList = append(peerList[:i], peerList[i+1:]...)
				break
			}
		}
		fmt.Println("Peer disconnected:", peerAddr)
	}()
	// Get requested file name
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading file name: ", err)
		return
	}
	fileName := string(buffer[:n])
	fileRequests <- fileName
}

func handleWebServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the P2P File Sharing Application\n")
		fmt.Fprintf(w, "You can use '/files' to see available files\n ")
		fmt.Fprintf(w, "You can use '/dowunload?file=filename'...the file with the specifed name will be downloaded if found\n ")
	})

	http.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		// Display a list of available files
		files, _ := filepath.Glob("*.*")
		fmt.Fprintf(w, "Available Files: %s", files)
	})
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		// Get the requested file name
		fileName := r.URL.Query().Get("file")
		fileRequests <- fileName

		fmt.Fprintf(w, "Downloading file: %s", fileName)
	})
	http.ListenAndServe("192.168.43.37:8080", nil)
}
