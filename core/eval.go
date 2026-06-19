package core

import (
	"errors"
	"io"
	"strconv"
	"time"
)

var RESP_NIL = []byte("$-1\r\n")

func evalPING(args []string, conn io.ReadWriter) error {
	var b []byte

	if len(args) >= 2 {
		return errors.New("ERR wrong number of arguments for 'ping' command")
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	_, err := conn.Write(b)
	return err
}

func evalSET(args []string, conn io.ReadWriter) error {
	if len(args) <= 1 {
		return errors.New("(error) ERR wrong number of arguments for 'set' command")
	}

	key, value := args[0], args[1]
	exDurationMs := int64(-1)

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return errors.New("(error) ERR syntax error")
			}
			exDurationSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return errors.New("(error) ERR value is not an integer or out of range")
			}

			exDurationMs = exDurationSec * 1000
		default:
			return errors.New("(error) ERR syntax error")
		}
	}

	Put(key, NewObj(value, exDurationMs))
	conn.Write([]byte("+OK\r\n"))
	return nil
}

func evalGET(args []string, conn io.ReadWriter) error {

	if len(args) != 1 {
		return errors.New("(error) ERR wrong number of arguments for 'get' command")
	}

	key := args[0]
	obj := Get(key)

	if obj == nil {
		conn.Write(RESP_NIL)
		return nil
	}

	if obj.ExpiresAt != -1 && obj.ExpiresAt < time.Now().UnixMilli() {
		conn.Write(RESP_NIL)
		return nil
	}

	conn.Write(Encode(obj.Value, false))
	return nil
}

func evalTTL(args []string, conn io.ReadWriter) error {

	if len(args) != 1 {
		return errors.New("(error) ERR wrong number of arguments for 'ttl' command")
	}

	obj := Get(args[0])

	if obj == nil {
		conn.Write([]byte(":-2\r\n"))
		return nil
	}

	if obj.ExpiresAt == -1 {
		conn.Write([]byte(":-1\r\n"))
		return nil
	}

	duration := obj.ExpiresAt - time.Now().UnixMilli()
	if duration < 0 {
		conn.Write([]byte(":-2\r\n"))
		return nil
	}

	conn.Write(Encode(int64(duration/1000), false))
	return nil
}

func EvalAndRespond(cmd *BedisCmd, conn io.ReadWriter) error {
	switch cmd.Cmd {
	case "PING":
		return evalPING(cmd.Args, conn)
	case "SET":
		return evalSET(cmd.Args, conn)
	case "GET":
		return evalGET(cmd.Args, conn)
	case "TTL":
		return evalTTL(cmd.Args, conn)
	default:
		return evalPING(cmd.Args, conn)
	}
}
