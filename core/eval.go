package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

var (
	RESP_NIL     = []byte("$-1\r\n")
	RESP_OK      = []byte("+OK\r\n")
	RESP_ONE     = []byte(":1\r\n")
	RESP_ZERO    = []byte(":0\r\n")
	RESP_MINUS_1 = []byte(":-1\r\n")
	RESP_MINUS_2 = []byte(":-2\r\n")
)

func evalPING(args []string) []byte {
	var b []byte

	if len(args) >= 2 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'ping' command"), false)
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	return b
}

func evalSET(args []string) []byte {
	if len(args) <= 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'set' command"), false)
	}

	key, value := args[0], args[1]
	oType, oEnc := getTypeEncoding(value)
	exDurationMs := int64(-1)

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return Encode(errors.New("(error) ERR syntax error"), false)
			}
			exDurationSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return Encode(errors.New("(error) ERR value is not an integer or out of range"), false)
			}

			exDurationMs = exDurationSec * 1000
		default:
			return Encode(errors.New("(error) ERR syntax error"), false)
		}
	}

	Put(key, NewObj(value, exDurationMs, oType, oEnc))
	return RESP_OK
}

func evalGET(args []string) []byte {

	if len(args) != 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'get' command"), false)
	}

	key := args[0]
	obj := Get(key)

	if obj == nil {
		return RESP_NIL
	}

	if obj.ExpiresAt != -1 && obj.ExpiresAt < time.Now().UnixMilli() {
		return RESP_NIL
	}

	return Encode(obj.Value, false)
}

func evalTTL(args []string) []byte {

	if len(args) != 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'ttl' command"), false)
	}

	obj := Get(args[0])

	if obj == nil {
		return RESP_MINUS_2
	}

	if obj.ExpiresAt == -1 {
		return RESP_MINUS_1
	}

	duration := obj.ExpiresAt - time.Now().UnixMilli()
	if duration < 0 {
		return RESP_MINUS_2
	}

	return Encode(int64(duration/1000), false)
}

func evalDEL(args []string) []byte {

	if len(args) < 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'del' command"), false)
	}

	countDel := 0
	for _, key := range args {
		if ok := Del(key); ok {
			countDel++
		}
	}

	return Encode(countDel, false)
}

func evalEXPIRE(args []string) []byte {

	if len(args) <= 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'expire' command"), false)
	}

	key := args[0]
	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("(error) ERR value is not an integer or out of range"), false)
	}

	obj := Get(key)

	if obj == nil {
		return RESP_ZERO
	}

	obj.ExpiresAt = time.Now().UnixMilli() + exDurationSec*1000

	return RESP_ONE
}

func evalBGWRITEAOF() []byte {
	go DumpAllAOF()
	return RESP_OK
}

func evalINCR(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'INCR' command"), false)
	}

	key := args[0]
	obj := Get(key)
	if obj == nil {
		obj = NewObj("0", -1, OBJ_TYPE_STRING, OBJ_ENCODING_INT)
		Put(key, obj)
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_STRING); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_INT); err != nil {
		return Encode(err, false)
	}

	i, _ := strconv.ParseInt(obj.Value.(string), 10, 64)
	i++
	obj.Value = strconv.FormatInt(i, 10)

	return Encode(i, false)	
}

func evalINFO() []byte {
	var info []byte
	buff := bytes.NewBuffer(info)

	buff.WriteString("# Keyspace\r\n")
	for range KeyStats {
		buff.WriteString(fmt.Sprintf("db0:keys=%d,expires=0,avg_ttl=0\r\n", KeyStats[KEY_METRIC]))
	}
	
	return Encode(buff.String(), false)
}

func EvalAndRespond(cmds BedisCmds, c io.ReadWriter) {
	var response []byte
	buf := bytes.NewBuffer(response)

	for _, cmd := range cmds {
		switch cmd.Cmd {
		case "PING", "ping":
			buf.Write(evalPING(cmd.Args))
		case "SET", "set":
			buf.Write(evalSET(cmd.Args))
		case "GET", "get":
			buf.Write(evalGET(cmd.Args))
		case "TTL", "ttl":
			buf.Write(evalTTL(cmd.Args))
		case "DEL", "del":
			buf.Write(evalDEL(cmd.Args))
		case "EXPIRE", "expire":
			buf.Write(evalEXPIRE(cmd.Args))
		case "BGWRITEAOF", "bgwriteaof":
			buf.Write(evalBGWRITEAOF())
		case "INCR", "incr":
			buf.Write(evalINCR(cmd.Args))
		case "INFO", "info":
			buf.Write(evalINFO())
		default:
			buf.Write(evalPING(cmd.Args))
		}
	}
	c.Write(buf.Bytes())
}
