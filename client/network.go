package main

import (
	"fmt"
	"net"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type connectionStatus struct {
	err error
}

const serverAddr = "127.0.0.1:3000"

// TODO: ON CONNECTION SEND A REGISTER MESSAGE WITH USERNAME AND MAP ADDR WITH USERNAME
// FETCH USERNAME FROM THE MAP AND SEND OTHER CLIENTS ON THE SERVER
// ----
// MAYBE THE USERNAME EXISTS... WHAT HAPPENS NEXT? MAYBE SERVER WILL SEND YOU SELECT A DIFFERNT ONE
// ORRRR IT GIVES YOU A  NAME LIKE USERNAME+RANDOM NUMBERS123
// LETS assume this wont happen i don't care, go away

func Connect(addr string, p *tea.Program) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		p.Send(connectionStatus{err: fmt.Errorf("resolve error: %w", err)})
		return
	}

	con, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		p.Send(connectionStatus{err: fmt.Errorf("dial error: %w", err)})
		return
	}

	p.Send(registerCon{con: con})
	p.Send(connectionStatus{err: nil})

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := con.Read(buf)
			if err != nil {
				break
			}
			if n < 1 {
				continue
			}
			splits := strings.Split(string(buf[:n-1]), ">>")
			from, payload := splits[0], splits[1]
			var fromType MessageFrom
			switch from {
			case "Room":
				fromType = FromRoom
			default:
				fromType = FromOther
			}
			p.Send(msgReceived{from: from, value: payload, fromType: fromType})
		}
	}()
}
