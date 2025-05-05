package handler
import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"github.com/shreyashghadge11/redis-go/redis"
)

func HandleCommand(redisCache *redis.Redis, args []string, conn net.Conn) {
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
			redisCache.AddToMultiCommand(strings.Join(args, " "))
			conn.Write([]byte("QUEUED\n"))
		} else {
			execute(cmd, args, conn, redisCache)
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
		handleSetCommand(redisCache, args, conn)
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

func handleSetCommand(redisCache *redis.Redis, args []string, conn net.Conn) {
	if len(args) < 3 {
		conn.Write([]byte("Invalid command\n"))
		return
	}
	key := strings.TrimSpace(args[1])
	value := strings.TrimSpace(args[2])
	redisCache.Set(key, value)

	if len(args) > 4 {
		if args[3] == "EX" {
			ttl, err := strconv.Atoi(strings.TrimSpace(args[4]))
			if err != nil {
				conn.Write([]byte("Invalid command\n"))
				return
			}
			redisCache.SetTTL(key, ttl)
		}
	}
	conn.Write([]byte("OK\n"))
}
