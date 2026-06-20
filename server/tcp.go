package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"syscall"
	"time"

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
		Cmd:  tokens[0],
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

var no_of_clients int = 0
var cronFrequency time.Duration = 1 * time.Second
var lastCronExecTime time.Time = time.Now()

func RunServer() error {
	log.Println("Connecting to bedis server on " + config.Host + ":" + strconv.Itoa(config.Port))

	max_clients := 20000

	// Events will hold all the sockets that are ready to be read from. If we do conn.Read() here then it will read something for sure
	events := make([]syscall.EpollEvent, max_clients)

	// This creates a socket for us to use
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(serverFD)

	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	// Binds the ip to the socket we created
	ip4 := net.ParseIP(config.Host)
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		return err
	}

	// start listening on the socket we just created
	if err = syscall.Listen(serverFD, max_clients); err != nil {
		return err
	}

	//this creates an epoll instance
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		return err
	}
	defer syscall.Close(epollFD)

	// This adds a epoll event that we want to moniter. EOLLIN means that when the socket is ready to be read from for the particular file descriptor 'serverFD' then we add it to the events list.
	socketServerEvent := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}

	// This registers the server with the epoll instance so that we can monitor it for events
	if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &socketServerEvent); err != nil {
		return err
	}

	for {

		// This is how we auto clean expired in a single treaded event loop. Everytime the conntrol flow reaches this part we do a clean up job if 1 second has passed
		if time.Now().After(lastCronExecTime.Add(cronFrequency)) {
			core.DeleteExpiredKeys()
			lastCronExecTime = time.Now()
		}

		// Gives the number of FD that are triggering event
		nevents, e := syscall.EpollWait(epollFD, events[:], -1)
		if e != nil {
			continue
		}

		for i := range nevents {

			// if the FD that is triggering is the server itself then that means we have a new connection that is trying to connect to server
			if events[i].Fd == int32(serverFD) {
				// accept the connection
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("error: ", err)
					continue
				}

				// increase the number of clients
				no_of_clients++
				syscall.SetNonblock(int(fd), true)

				// Add the new connection to the monitor list
				socketClientEvent := syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(fd),
				}

				if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &socketClientEvent); err != nil {
					log.Fatal(err)
				}
			} else {
				comm := core.FDcomm{Fd: int(events[i].Fd)}
				cmd, err := readCommand(comm)
				if err != nil {
					syscall.Close(int(events[i].Fd))
					no_of_clients--
					continue
				}

				respond(cmd, comm)
			}
		}
	}
}
