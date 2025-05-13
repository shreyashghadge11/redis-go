package redis

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

type redisObject struct {
	datatype reflect.Type
	value    interface{}
}

type multiCommand struct {
	commands         []string
	isMultiCommandOn bool
}

type Redis struct {
	dict         map[string]redisObject
	mu           sync.RWMutex
	ttl          map[string]int64
	multiCommand map[string]multiCommand
}

func NewRedis() *Redis {
	return &Redis{
		dict:         make(map[string]redisObject),
		ttl:          make(map[string]int64),
		multiCommand: make(map[string]multiCommand),
	}
}

func (r *Redis) SetTTL(key string, ttl int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ttl[key] = time.Now().Add(time.Duration(ttl) * time.Second).Unix()
}

func (r *Redis) Set(key string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.dict[key] = redisObject{
		datatype: reflect.TypeOf(value),
		value:    value,
	}
}

func (r *Redis) Get(key string) interface{} {
	if _, ok := r.dict[key]; !ok {
		return nil
	}
	return r.dict[key].value
}

func (r *Redis) Del(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.dict[key]; ok {

		delete(r.dict, key)
		return true
	}

	return false
}

func (r *Redis) Exists(key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.dict[key]
	return exists
}

func (r *Redis) FlushAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.dict = make(map[string]redisObject)
}

func (r *Redis) Increment(key string, val float64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	obj, exists := r.dict[key]
	if !exists {
		// If key does not exist, create it with initial value
		r.dict[key] = redisObject{
			datatype: reflect.TypeOf(val),
			value:    val,
		}
		return true, nil
	}

	switch v := obj.value.(type) {
	case int:
		newVal := float64(v) + val
		obj.value = newVal
		obj.datatype = reflect.TypeOf(newVal)
	case float64:
		obj.value = v + val
	default:
		return false, fmt.Errorf("ERR value is not a number")
	}

	// Store updated object
	r.dict[key] = obj
	return true, nil
}

func (r *Redis) StartMultiCmds(connKey string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.multiCommand[connKey] = multiCommand{
		commands:         []string{},
		isMultiCommandOn: true,
	}
}

func (r *Redis) MultiCmdStatus(connKey string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if cmd, ok := r.multiCommand[connKey]; ok {
		return cmd.isMultiCommandOn
	}
	return false
}

func (r *Redis) GetMultiCommands(connKey string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if cmd, ok := r.multiCommand[connKey]; ok {
		return cmd.commands
	}
	return nil
}

func (r *Redis) AddToMultiCommand(connKey string, command string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cmd, ok := r.multiCommand[connKey]; ok {
		cmd.commands = append(cmd.commands, command)
		r.multiCommand[command] = cmd
	}
}

func (r *Redis) Discard(connKey string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.multiCommand[connKey]; ok {
		delete(r.multiCommand, connKey)
	}
}
