package gohm

import (
	"github.com/garyburd/redigo/redis"
)

type Command interface {
	Call(prefix string, conn redis.Conn) (string, error)
	Clean() error
}

type SimpleCommand struct {
	Key string
}

func NewSimpleCommand(key string) Command {
	return &SimpleCommand{
		Key: key,
	}
}

func (c *SimpleCommand) Call(prefix string, conn redis.Conn) (string, error) {
	return c.Key, nil
}

func (c *SimpleCommand) Clean() error {
	return nil
}

type CompositeCommand struct {
	Operation string
	Arguments []Command
	Key       string
	KeyConn   redis.Conn
}

func NewCommandWithSubs(op string, argument ...Command) Command {
	if len(argument) == 1 {
		return argument[0]
	}
	command := &CompositeCommand{
		Operation: op,
		Arguments: make([]Command, len(argument)),
	}
	copy(command.Arguments, argument)
	return command
}

func (c *CompositeCommand) Call(prefix string, conn redis.Conn) (string, error) {
	suffix, err := generateRandomHexString(32)
	if err != nil {
		return "", err
	}
	c.Key = connectKeys(prefix, suffix)
	params := make([]interface{}, len(c.Arguments)+1)
	params[0] = c.Key
	for i := range c.Arguments {
		params[i+1], err = c.Arguments[i].Call(prefix, conn)
		if err != nil {
			return "", err
		}
	}
	_, err = conn.Do(c.Operation, params...)
	if err != nil {
		return "", err
	}
	c.KeyConn = conn
	return c.Key, nil
}

func (c *CompositeCommand) Clean() error {
	if len(c.Key) > 0 && c.KeyConn != nil {
		_, err := c.KeyConn.Do("DEL", c.Key)
		if err != nil {
			return err
		}
		c.Key = ""
		c.KeyConn = nil
	}
	for i := range c.Arguments {
		err := c.Arguments[i].Clean()
		if err != nil {
			return err
		}
	}
	return nil
}

func NewCommand(op string, keys ...string) Command {
	commands := make([]Command, len(keys))
	for i := range keys {
		commands[i] = NewSimpleCommand(keys[i])
	}
	return NewCommandWithSubs(op, commands...)
}
