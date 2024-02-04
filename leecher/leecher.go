package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/anacrolix/torrent"
	// "github.com/anacrolix/torrent"
)

func main() {
	fmt.Println("Starting the application...")
	// 10.5.226.133...mine ip undet staff wifi
	// Connect to this socket
	conn, err := net.Dial("tcp", "192.168.8.104:1234") //22
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	fmt.Println("Connected to server", conn.RemoteAddr().String())

	// go wait()

	go handleFileReceive()

	for {
		// Read message from user
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			log.Fatal(err)
		}
		fmt.Print("Message from server: ", string(buf[:n]))
	}

}

func handleFileReceive() {
	// Listen for incoming connections
	ln, err := net.Listen("tcp", "192.168.8.103:8080")
	if err != nil {
		fmt.Println("Error listening: ", err)
		os.Exit(1)
		return
	}

	// Close the listener when the application closes
	defer ln.Close()

	fmt.Println("Listening on 192.168.43.37:8080")

	for {
		// Listen for an incoming connection
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err)
			continue
		}

		var fileName string
		// var fileContent []byte
		_, err = fmt.Fscanln(conn, &fileName)
		if err != nil {
			fmt.Println("Error receiving file name:", err.Error())
			conn.Close()
			continue
		}

		fmt.Println("file name", fileName)

		fmt.Println("Receiving file:", fileName)
		fileContent := make([]byte, 1024)
		n, err := conn.Read(fileContent)
		if err != nil {
			fmt.Println("Error receiving file content:", err.Error())
		}

		fmt.Println(string(fileContent[:n]))

		// create new file and save content
		newFile, err := os.Create(fileName)
		if err != nil {
			fmt.Println("Error creating file:", err.Error())
			conn.Close()
			continue
		}
		fmt.Println("File created")

		data := []byte(string(fileContent[:n]))

		_, err = newFile.Write(data)
		if err != nil {
			fmt.Println("Error writing file content:", err.Error())
			conn.Close()
			continue
		}
		newFile.Close()
		fmt.Println("File content written")

		fileExtension := filepath.Ext(fileName)
		fmt.Println("file ext", fileExtension)

		if fileExtension == ".torrent" {
			fmt.Println("file with", fileExtension, "Torrent file selected")
			go downloadTorrent(fileName)
		} else {
			fmt.Println("file with", fileExtension, "Regular file selected")
			// downloadRegularFile(fileName)
		}

		conn.Close()

	}

}

func downloadTorrent(fileName string) {
	client, err := torrent.NewClient(nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer client.Close()

	// Parse the torrent file
	t, err := client.AddTorrentFromFile(fileName)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Create a channel to receive the download progress
	progress := make(chan int)

	// Create a goroutine to monitor the download progress
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case p := <-progress:
				fmt.Printf("Downloaded %d%%\n", p)
			case <-t.GotInfo():
				progress <- int(t.BytesCompleted() * 100 / t.Info().TotalLength())
			}
		}
	}()

	fmt.Println("Downloading file: ", t.Info().Name)
	fmt.Println("File size: ", t.Info().TotalLength()/1024/1024, "MB")
	fmt.Println("Number of pieces: ", t.NumPieces())
	fmt.Println("Number of files: ", len(t.Files()))

	// fmt.Println("Number of peers: ", len(t.Peers()))
	// fmt.Println("Number of seeds: ", len(t.Seeds()))
	// fmt.Println("Number of leechers: ", len(t.Leechers()))
	// fmt.Println("Download speed: ", t.DownloadRate())

	// Start the download
	t.DownloadAll()

	// Wait for all goroutines to finish
	wg.Wait()

	// Print the download statistics
	fmt.Printf("Downloaded %d bytes\n", t.BytesCompleted())
	fmt.Printf("Download speed: %d KB/s\n", t.Stats())
	// fmt.Printf("Upload speed: %d KB/s\n", t.UploadRate()/1024)

	// Open the file
	file, err := os.Open(t.Info().Name)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	// Read the file
	data := make([]byte, t.Info().TotalLength())
	_, err = file.Read(data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Print the file contents
	fmt.Println(string(data))
}
