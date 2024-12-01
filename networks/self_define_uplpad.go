package networks

import (
	"blockEmulator/params"
	"golang.org/x/time/rate"
	"log"
	"net"
)

var (
	Upload *rate.Limiter
)

func InitNetworkSpecial() {
	t := params.Bandwidth * (params.MigrateNodeNum / params.ZoneSize)
	Upload = rate.NewLimiter(rate.Limit(t), t)
}

func TcpDialSpecial(context []byte, addr string) {
	go func() {
		connMaplock.Lock()
		defer connMaplock.Unlock()

		var err error
		var conn net.Conn // Define conn here

		// if this connection is not built, build it.
		if c, ok := connectionPool[addr]; ok {
			if tcpConn, tcpOk := c.(*net.TCPConn); tcpOk {
				if err := tcpConn.SetKeepAlive(true); err != nil {
					delete(connectionPool, addr) // Remove if not alive
					conn, err = net.Dial("tcp", addr)
					if err != nil {
						log.Println("Reconnect error", err)
						return
					}
					connectionPool[addr] = conn
					go ReadFromConn(addr) // Start reading from new connection
				} else {
					conn = c // Use the existing connection
				}
			}
		} else {
			conn, err = net.Dial("tcp", addr)
			if err != nil {
				log.Println("Connect error", err)
				return
			}
			connectionPool[addr] = conn
			go ReadFromConn(addr) // Start reading from new connection
		}

		writeToConn(append(context, Delimit), conn, rateLimiterUpload)
	}()
}
