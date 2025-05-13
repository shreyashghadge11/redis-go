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
	connAddr := conn.RemoteAddr().String()

	switch cmd {
	case "MULTI":
		redisCache.StartMultiCmds(connAddr)
		conn.Write([]byte("OK\n"))
	case "EXEC":
		isMultiOn := redisCache.MultiCmdStatus(connAddr)
		if isMultiOn {
			commands := redisCache.GetMultiCommands(connAddr)
			executeMultiCommands(commands, conn, redisCache)
			redisCache.Discard(connAddr)
			conn.Write([]byte("OK\n"))
		} else {
			conn.Write([]byte("Error: MULTI not set\n"))
		}
	case "DISCARD":
		redisCache.Discard(connAddr)
		conn.Write([]byte("OK\n"))
	default:
		if redisCache.MultiCmdStatus(connAddr) {
			redisCache.AddToMultiCommand(connAddr, strings.TrimSpace(strings.Join(args, " ")))
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
	switch cmd {
	case "SET":
		handleSetCommand(redisCache, args, conn)
	case "GET":
		handleGetCommand(redisCache, args, conn)
	case "EXISTS":
		handleExistsCommand(redisCache, args, conn)
	case "DEL":
		handleDeleteCommand(redisCache, args, conn)
	case "INCR":
		handleIncrCommand(redisCache, args, conn)
	case "INCRBY":
		handleIncrByCommand(redisCache, args, conn)
	case "DECR":
		handleIncrCommand(redisCache, args, conn)
	case "DECRBY":
		handleIncrByCommand(redisCache, args, conn)
	case "PING":
		conn.Write([]byte("PONG\n"))
	case "FLUSHALL":
		redisCache.FlushAll()
		conn.Write([]byte("OK\n"))
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
	val := strings.TrimSpace(args[2])

	if intVal, err := strconv.Atoi(val); err == nil {
		redisCache.Set(key, intVal)
	} else if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
		redisCache.Set(key, floatVal)
	} else {
		redisCache.Set(key, val) // Default to string
	}

	if len(args) > 4 {
		if args[3] == "EX" {
			ttl, err := strconv.Atoi(strings.TrimSpace(args[4]))
			if err != nil {
				conn.Write([]byte("Invalid command\n"))
				return
			}
			ttl = ttl * 60 // Convert to seconds
			redisCache.SetTTL(key, ttl)
		}
	} else {
		redisCache.SetTTL(key, 600) // Default TTL
	}
	conn.Write([]byte("OK\n"))
}

func handleGetCommand(redisCache *redis.Redis, args []string, conn net.Conn) {
	if len(args) < 2 {
		conn.Write([]byte("Invalid command\n"))
		return
	}
	key := strings.TrimSpace(args[1])
	value := redisCache.Get(key)
	if value == nil {
		conn.Write([]byte("nil\n"))
	} else {
		conn.Write([]byte(fmt.Sprintf("%v\n", value)))
	}
}

func handleExistsCommand(redisCache *redis.Redis, args []string, conn net.Conn) {
	if len(args) < 2 {
		conn.Write([]byte("Invalid command\n"))
		return
	}
	key := strings.TrimSpace(args[1])
	if redisCache.Get(key) != nil {
		conn.Write([]byte("1\n"))
	} else {
		conn.Write([]byte("0\n"))
	}
}

func handleDeleteCommand(redisCache *redis.Redis, args []string, conn net.Conn) {
	if len(args) < 2 {
		conn.Write([]byte("Invalid command\n"))
		return
	}
	key := strings.TrimSpace(args[1])
	if redisCache.Del(key) {
		conn.Write([]byte("1\n"))
	} else {
		conn.Write([]byte("0\n"))
	}
}

func handleIncrCommand(redisCache *redis.Redis, args []string, conn net.Conn) {
	if len(args) < 2 {
		conn.Write([]byte("Invalid command\n"))
		return
	}
	key := strings.TrimSpace(args[1])
	if _, err := redisCache.Increment(key, 1.0); err == nil {
		conn.Write([]byte("OK\n"))
	} else {
		conn.Write([]byte("Error: " + err.Error() + "\n"))	
	}
}

func handleIncrByCommand(redisCache *redis.Redis, args []string, conn net.Conn) {
	if len(args) < 3 {
		conn.Write([]byte("Invalid command\n"))
		return
	}
	key := strings.TrimSpace(args[1])
	increment, err := strconv.ParseFloat(strings.TrimSpace(args[2]), 64)
	if err != nil {
		conn.Write([]byte("Invalid command\n"))
		return
	}
	if _, err := redisCache.Increment(key, float64(increment)); err == nil {
		conn.Write([]byte("OK\n"))
	} else {
		conn.Write([]byte("Error: " + err.Error() + "\n"))	
	}
}