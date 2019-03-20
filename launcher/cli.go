package launcher

import (
	"bufio"
	"log"
	"net"
	"os"
)

func runCli(config *Config) error {
	socket, err := net.Dial("tcp", "127.0.0.1:21026")
	defer socket.Close()
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Print("Connected")
	rc := make(chan int)
	wc := make(chan int)
	go func() {
		writer := bufio.NewWriter(socket)
		writer.ReadFrom(os.Stdin)
		wc <- 0
	}()

	go func() {
		reader := bufio.NewReader(socket)
		reader.WriteTo(os.Stdout)
		rc <- 0
	}()
	select {
	case <-rc:
	case <-wc:
	}
	return nil
}
