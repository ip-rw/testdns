package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/phuslu/fastdns"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"strings"
	"sync"
	"testdns/pkg/dns"
	"time"
)

func stdinReader() chan string {
	outchan := make(chan string, *workers*5)
	scan := bufio.NewScanner(os.Stdin)
	go func() {
		for scan.Scan() {
			if scan.Err() != nil {
				fmt.Println("Error reading stdin:", scan.Err())
				return
			}
			outchan <- scan.Text()
		}
		close(outchan)
	}()
	return outchan
}

func worker(hostname string, serverChan chan string, resultChan chan *dns.Result) {
	req, resp := fastdns.AcquireMessage(), fastdns.AcquireMessage()
	for server := range serverChan {
		resultChan <- dns.QueryReuse(hostname, server, req, resp, *timeout)
	}
	fastdns.ReleaseMessage(req)
	fastdns.ReleaseMessage(resp)
}

var (
	workers  = flag.Int("w", 25, "workers")
	hostname = flag.String("n", "test-12-34-56-78.nip.io", "hostname to resolve")
	invert   = flag.Bool("v", false, "invert output, show misbehaving servers")
	silent   = flag.Bool("s", true, "don't show response time.")
	trusted  = flag.String("r", "1.1.1.1:53", "trusted resolver")
	timeout  = flag.Duration("t", 2*time.Second, "timeout")
)

func init() {
	flag.Parse()
	if *hostname == "" {
		logrus.Errorln("please specify hostname")
		os.Exit(1)
	}
	if *silent {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	trusted := dns.Query(*hostname, *trusted, *timeout)
	if trusted.Error != nil {
		logrus.Infof("error resolving %s (%s) - using nxdomain", *hostname, trusted.Error)
	}
	wg := &sync.WaitGroup{}
	serverChan := make(chan string, *workers*2)
	resultChan := make(chan *dns.Result, *workers*2)

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			worker(*hostname, serverChan, resultChan)
			wg.Done()
		}()
	}

	go func() {
		for line := range stdinReader() {
			serverChan <- line
		}
		close(serverChan)
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		if result.Error != nil {
			//logrus.Debugln(result.Server, result.Error)
			continue
		}
		display := result.Matches(trusted)
		if !display {
			if result.Error == nil {
				logrus.Warnf("%s didn't match (%s != %s)", formatResult(result), formatIPs(trusted.Answer), formatIPs(result.Answer))
			}
		}
		if *invert {
			display = !display
		}
		if display {
			fmt.Println(formatResult(result))
		}
	}
}

func formatIPs(ips []net.IP) string {
	s := strings.Builder{}
	for _, ip := range ips {
		s.WriteString(ip.String() + ", ")
	}
	return strings.Trim(s.String(), " ,")
}

func formatResult(result *dns.Result) string {
	if *silent {
		return fmt.Sprintf("%s\t%d", result.Server, result.Time.Milliseconds())
	} else {
		return result.Server
	}
}
