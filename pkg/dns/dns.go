package dns

import (
	"github.com/phuslu/fastdns"
	"net"
	"strings"
	"time"
)

type Result struct {
	Server   string
	Question string
	Answer   []net.IP
	Time     time.Duration
	Error    error
}

func (r *Result) Matches(r2 *Result) bool {
	if len(r.Answer) == 0 && len(r2.Answer) == 0 {
		return true
	}
	for _, ip := range r.Answer {
		for _, ip2 := range r2.Answer {
			if ip.Equal(ip2) {
				return true
			}
		}
	}
	return false
}

func Query(domain, server string, timeout time.Duration) (result *Result) {
	req, resp := fastdns.AcquireMessage(), fastdns.AcquireMessage()
	defer fastdns.ReleaseMessage(req)
	defer fastdns.ReleaseMessage(resp)
	return QueryReuse(domain, server, req, resp, timeout)
}

func QueryReuse(domain, server string, req, resp *fastdns.Message, timeout time.Duration) (result *Result) {
	result = &Result{
		Server:   server,
		Question: domain,
	}

	defer func() {
		if r := recover(); r != nil {
			result.Error = r.(error)
		}
	}()

	req.SetRequestQustion(domain, fastdns.TypeA, fastdns.ClassINET)

	if !strings.Contains(server, ":") {
		server = server + ":53"
	}
	serverAddr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		result.Error = err
	}
	start := time.Now()

	err = Exchange(serverAddr, req, resp, timeout)
	result.Time = time.Now().Sub(start)
	if err != nil {
		result.Error = err
		return result
	}
	_ = resp.Walk(func(name []byte, typ fastdns.Type, class fastdns.Class, ttl uint32, data []byte) bool {
		switch typ {
		case fastdns.TypeA, fastdns.TypeAAAA:
			result.Answer = append(result.Answer, data)
		}
		return true
	})
	return result
}
