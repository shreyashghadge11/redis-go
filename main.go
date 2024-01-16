package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"

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
		cmd := strings.ToUpper(args[0])

		switch cmd {
		case "MULTI":
			conn.Write([]byte("OK\n"))
			redisCache.Multi()
		case "EXEC":
			isMultiOn := redisCache.Status()
			if isMultiOn {
				conn.Write([]byte("OK\n"))
				commands := redisCache.Exec()
				executeMultiCommands(commands, conn, redisCache)

			} else {
				conn.Write([]byte("Error: MULTI not set\n"))
			}
		case "DISCARD":
			redisCache.Discard()
			conn.Write([]byte("OK\n"))
		default:
			if redisCache.Status() {
				redisCache.AddToMultiCommand(message)
				conn.Write([]byte("QUEUED\n"))
			} else {
				execute(cmd, args, conn, redisCache)
			}
		}

	}
}

func executeMultiCommands(multiCommand []string, conn net.Conn, redisCache *redis.Redis) {
	for _, command := range multiCommand {
		args := strings.Split(strings.TrimSpace(command), " ")
		cmd := strings.ToUpper(args[0])
		execute(cmd, args, conn, redisCache)
	}
}

func execute(cmd string, args []string, conn net.Conn, redisCache *redis.Redis) {
	key := strings.TrimSpace(args[1])
	switch cmd {
	case "SET":
		if len(args) < 3 {
			conn.Write([]byte("Invalid command\n"))
			return
		}
		val := strings.TrimSpace(args[2])
		Value := interface{}(val)
		redisCache.Set(key, Value)
		conn.Write([]byte("OK\n"))
	case "GET":
		val := redisCache.Get(key)
		if val == nil {
			conn.Write([]byte("(nil)\n"))
		} else {
			conn.Write([]byte(fmt.Sprintf("%v\n", val)))
		}
	case "DEL":
		if redisCache.Del(key) {
			conn.Write([]byte("1\n"))
		} else {
			conn.Write([]byte("0\n"))
		}
	case "INCR":
		if redisCache.Increment(key) {
			conn.Write([]byte("1\n"))
		} else {
			conn.Write([]byte("0\n"))
		}
	case "INCRBY":
		if len(args) < 3 {
			conn.Write([]byte("Invalid command\n"))
			return
		}
		increment, err := strconv.ParseFloat(strings.TrimSpace(args[2]), 64)
		if err != nil {
			conn.Write([]byte("Invalid command\n"))
			return
		}
		if redisCache.IncrementBy(key, float64(increment)) {
			conn.Write([]byte("1\n"))
		} else {
			conn.Write([]byte("0\n"))
		}
	default:
		conn.Write([]byte("Invalid command\n"))
	}
}
