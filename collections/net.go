package collections

import (
	"fmt"
	"log"

	"github.com/shirou/gopsutil/v3/net"
)

type NetworkData struct {
	Proto      string `json:"protocol"`
	LocalAddr  string `json:"local_address"`
	RemoteAddr string `json:"remote_address"`
	State      string `json:"state"`
	Pid        int32  `json:"pid"`
}

func NetCollect() []NetworkData {
	var list []NetworkData

	conns, err := net.Connections("inet")
	if err != nil {
		log.Printf("[-] Error retrieving network connections: %v", err)
		return list
	}

	for _, c := range conns {
		lAddr := fmt.Sprintf("%s:%d", c.Laddr.IP, c.Laddr.Port)
		rAddr := ""
		if c.Raddr.IP != "" {
			rAddr = fmt.Sprintf("%s:%d", c.Raddr.IP, c.Raddr.Port)
		}

		proto := "tcp"
		if c.Type == 2 {
			proto = "udp"
		}

		list = append(list, NetworkData{
			Proto:      proto,
			LocalAddr:  lAddr,
			RemoteAddr: rAddr,
			State:      c.Status,
			Pid:        c.Pid,
		})
	}

	return list
}
