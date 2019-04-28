package cli

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

func copyWorker(dst io.Writer, src io.Reader, doneCh chan<- bool) {
	io.Copy(dst, src)
	doneCh <- true
}

type ScreepsCLI struct {
	conn        net.Conn
	buffer      []string
	host        string
	port        int16
	readCh      chan string
	stopReadCh  chan bool
	WelcomeText string
}

func NewScreepsCLI(host string, port int16) *ScreepsCLI {
	return &ScreepsCLI{
		host: host,
		port: port,
	}
}

func (s *ScreepsCLI) Start() error {
	conn, err := net.Dial("tcp4", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		log.Println("dial:", err)
		return err
	}
	s.conn = conn
	s.readCh = make(chan string)
	s.stopReadCh = make(chan bool)

	go func(outCh chan<- string) {
		reader := bufio.NewReader(s.conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if len(line) > 0 {
				line = strings.TrimPrefix(line, "< ")
				outCh <- line
			}
		}
	}(s.readCh)
	buffer := ""
loop:
	for {
		select {
		case line := <-s.readCh:
			buffer = buffer + line
		case <-time.After(200 * time.Millisecond):
			break loop
		}
	}
	s.WelcomeText = buffer
	return nil
}

func (s *ScreepsCLI) Stop() {

}

func (s *ScreepsCLI) Command(cmd string) string {
	if len(cmd) > 0 {
		s.conn.Write([]byte(fmt.Sprintf("%s\n", cmd)))
	}
	buffer := make([]string, 0)
	log.Println("first sel")
	select {
	case line := <-s.readCh:
		buffer = append(buffer, line)
	case <-time.After(5 * time.Second):
		buffer = append(buffer, "Timeout Waiting for response")
		return strings.Join(buffer, "")
	}
loop:
	for {
		select {
		case line := <-s.readCh:
			buffer = append(buffer, line)
		case <-time.After(200 * time.Millisecond):
			break loop
		}
	}
	return strings.Join(buffer, "")
}
