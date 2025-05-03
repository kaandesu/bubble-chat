package main

import (
	"log"
	"net"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const serverAddr = "127.0.0.1:3000"

func Connect(addr string, p *tea.Program) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	temp := make(chan struct{})

	con, err := net.DialTCP("tcp", nil, tcpAddr)
	if err == nil {
		p.Send(registerCon{
			con: con,
		})
	}
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
	<-temp
}
