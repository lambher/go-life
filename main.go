package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	Message       = "Pong"
	StopCharacter = "\r\n\r\n"
)

type Cell struct {
	Alive bool
	Time  time.Time
}

var sizeX int
var sizeY int
var data [][]Cell

func main() {
	sizeX = 20
	sizeY = 20

	port := 3333

	go SocketServer(port)

	data = make([][]Cell, sizeX)
	for i, _ := range data {
		data[i] = make([]Cell, sizeY)
	}

	data[2][2].Alive = true
	data[2][2+1].Alive = true
	data[2+1][2].Alive = true

	ticker := time.NewTicker(1000 * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			//fmt.Println("Tick at", t)
			data = gameOfLife(data)
		}
	}

}

func display(data [][]Cell) {
	cmd := exec.Command("clear") //Linux example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()

	for x, _ := range data {
		for y, _ := range data[x] {
			if data[x][y].Alive {
				fmt.Print("X")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Print("\n")
	}
}

func gameOfLife(data [][]Cell) [][]Cell {
	copyData := copyData(data)

	for x, _ := range data {
		for y := range data[x] {
			data[x][y].Alive = isAlive(copyData, x, y)
		}
	}
	return data
}

func isAlive(data [][]Cell, x, y int) bool {
	number := getNumberLivingCells(data, x, y)
	if number == 3 {
		return true
	} else if number > 3 || number < 2 {
		return false
	}
	return data[x][y].Alive
}

func getNumberLivingCells(data [][]Cell, x, y int) int {
	number := 0

	if x > 0 && data[x-1][y].Alive {
		number++
	}
	if x > 0 && y > 0 && data[x-1][y-1].Alive {
		number++
	}
	if y > 0 && data[x][y-1].Alive {
		number++
	}
	if x < sizeX-1 && y > 0 && data[x+1][y-1].Alive {
		number++
	}
	if x < sizeX-1 && data[x+1][y].Alive {
		number++
	}
	if x < sizeX-1 && y < sizeY-1 && data[x+1][y+1].Alive {
		number++
	}
	if y < sizeY-1 && data[x][y+1].Alive {
		number++
	}
	if x > 0 && y < sizeY-1 && data[x-1][y+1].Alive {
		number++
	}
	return number
}

func copyData(data [][]Cell) [][]Cell {
	copyData := make([][]Cell, sizeX)
	for x, _ := range data {
		copyData[x] = make([]Cell, sizeY)
		for y := range data[x] {
			copyData[x][y] = data[x][y]
		}
	}

	return copyData
}

func SocketServer(port int) {
	listen, err := net.Listen("tcp4", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("Socket listen port %d failed,%s", port, err)
		os.Exit(1)
	}

	defer listen.Close()
	log.Printf("Begin listen port: %d", port)
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}
		fmt.Println("connexion")
		go handler(conn)
	}

}

func handler(conn net.Conn) {

	defer conn.Close()

	var (
		buf = make([]byte, 1024*48)
		r   = bufio.NewReader(conn)
		w   = bufio.NewWriter(conn)
	)

ILOOP:
	for {
		n, err := r.Read(buf)
		d := string(buf[:n])

		switch err {
		case io.EOF:
			break ILOOP
		case nil:
			//log.Println("Receive:", d)
			if isTransportOver(d) {
				break ILOOP
			}
			messages := readData(d)
			for _, message := range messages {
				ds := strings.Split(message, " ")
				length := len(ds)
				if len(ds) > 0 {
					if ds[0] == "ADD" {
						parseAdd(ds, length)
					}
				}

				if message == "GET MAP" {
					sendMap(w)
				}
			}


		default:
			log.Printf("Receive data failed:%s", err)
			return
		}

	}
	w.Write([]byte(Message))
	w.Flush()
	log.Printf("Send: %s", Message)

}

func readData(data string) []string {
	messages := make([]string, 0)
	strs := strings.Split(data, ";")
	for _, str := range strs {
		if str != ";" {
			messages = append(messages, str)
		}
	}

	return messages
}

func parseAdd(ds []string, length int) {
	if length > 2 {
		x, err := strconv.Atoi(ds[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		y, err := strconv.Atoi(ds[2])
		if err != nil {
			fmt.Println(err)
			return
		}

		if x < sizeX && y < sizeY && x >= 0 && y >= 0 {
			data[x][y].Alive = true
			fmt.Println("ADD", x, y)
		}
	}
}

func sendMap(w *bufio.Writer) {
	message, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		fmt.Println(err)
	}
	_, err = w.Write([]byte(";"))
	if err != nil {
		fmt.Println(err)
	}

	err = w.Flush()
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(string(message))
}

func isTransportOver(data string) (over bool) {
	over = strings.HasSuffix(data, "\r\n\r\n")
	return
}
