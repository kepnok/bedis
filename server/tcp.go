package server

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/kepnok/bedis/config"
	"github.com/kepnok/bedis/core"
)


func readCommand(conn io.ReadWriter) (*core.BedisCmd, error) {
	buf := make([]byte, 512)
	
	n, err := conn.Read(buf[:])
	if err != nil {
		return nil, err
	}

	tokens, err := core.ParseCmd(buf[:n])
	if err != nil {
		return nil, err
	}

	return &core.BedisCmd{
		Cmd: tokens[0],
		Args: tokens[1:],
	}, nil
}

func respondErr(err error, conn io.ReadWriter) {
	conn.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respond(cmd *core.BedisCmd, conn io.ReadWriter) error {
	err := core.EvalAndRespond(cmd, conn)
	if err != nil {
		respondErr(err, conn)
	}
	return nil
}

func RunServer() {
	log.Println("Connecting to bedis server on " + config.Host + ":" + string(config.Port))

	lsnr, err := net.Listen("tcp", config.Host + ":" + string(config.Port))
	if err != nil {
		panic(err)
	}

	for {
		conn, err := lsnr.Accept()
		if err != nil {
			panic(err)
		}

		for {
			cmd, err := readCommand(conn)
			if err != nil {
				conn.Close()

				log.Println("Connection closed ", conn.RemoteAddr())
				log.Println("err: ", err)

				break;
			}

			err = respond(cmd, conn)
			if err != nil {
				log.Println("err writing response: ", err)
			}
		}
	}

}