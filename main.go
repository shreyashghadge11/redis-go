package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	handler "github.com/shreyashghadge11/redis-go/cmd_handler"
	"github.com/shreyashghadge11/redis-go/redis"
)

func main() {
	redisCache := redis.NewRedis()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()

	fmt.Println("Listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		go handleConnection(conn, redisCache)
	}
}

func handleConnection(conn net.Conn, redisCache *redis.Redis) {
	defer conn.Close()

	fmt.Println("Handling connection from:", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			return
		}

		fmt.Println("Received message:", message)
		args := strings.Split(strings.TrimSpace(message), " ")

		if len(args) == 0 {
			conn.Write([]byte("Invalid command\n"))
			continue
		}

		handler.HandleCommand(redisCache, args, conn)
	}
}

func ExpiredKeyCleanup(redisCache *redis.Redis) {
	for {
		time.Sleep(60 * time.Second)
	}
}
