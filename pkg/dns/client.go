package dns

import (
	"github.com/phuslu/fastdns"
	"net"
	"os"
	"time"
)

// Exchange executes a single DNS transaction, returning
// a Response for the provided Request.
func Exchange(serverAddr *net.UDPAddr, req, resp *fastdns.Message, timeout time.Duration) (err error) {
	err = exchange(serverAddr, req, resp, timeout)
	if err != nil && os.IsTimeout(err) {
		err = exchange(serverAddr, req, resp, timeout)
	}
	return err
}

func exchange(serverAddr *net.UDPAddr, req, resp *fastdns.Message, timeout time.Duration) error {
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if timeout > 0 {
	}
	_, err = conn.Write(req.Raw)
	if err != nil {
		conn.Close()
		conn.SetWriteDeadline(time.Now().Add(timeout))
		if err != nil {
			return err
		}
		if _, err = conn.Write(req.Raw); err != nil {
			return err
		}
	}

	if timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(timeout))
	}
	resp.Raw = resp.Raw[:cap(resp.Raw)]
	n, err := conn.Read(resp.Raw)
	if err == nil {
		resp.Raw = resp.Raw[:n]
		err = fastdns.ParseMessage(resp, resp.Raw, false)
	}
	return err
}
