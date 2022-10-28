package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type DNSHandler struct {
	upstreams []string
	client    *dns.Client
}

func NewDNSHandler() (*DNSHandler, error) {
	timeout := 4

	handler := &DNSHandler{}

	if upst := os.Getenv("DNS_UPSTREAMS"); len(upst) > 10 {
		handler.upstreams = strings.Split(upst, ",")
	}

	if handler.upstreams == nil {
		cfg, err := dns.ClientConfigFromFile("/etc/resolv.conf")
		if err != nil {
			return nil, err
		}

		timeout = cfg.Timeout

		for _, s := range cfg.Servers {
			ip := net.ParseIP(s)

			if ip != nil && ip.To4() != nil {
				handler.upstreams = append(handler.upstreams, fmt.Sprintf("%s:%s", s, cfg.Port))
			} else {
				handler.upstreams = append(handler.upstreams, fmt.Sprintf("[%s]:%s", s, cfg.Port))
			}

		}
	}

	if handler.upstreams == nil {
		handler.upstreams = []string{
			// cloudflare
			"1.0.0.1:53",
			"1.1.1.1:53",

			// google
			"8.8.4.4:53",
			"8.8.8.8:53",
		}
	}

	clientTimeout := time.Duration(timeout) * time.Second

	handler.client = &dns.Client{
		Net:     "udp",
		Timeout: clientTimeout,
		Dialer: &net.Dialer{
			Timeout:   clientTimeout,
			LocalAddr: nil,
		},
	}

	return handler, nil
}

func (handler *DNSHandler) Query(ctx context.Context, binary []byte) ([]byte, error) {
	msg := new(dns.Msg)
	if err := msg.Unpack(binary); err != nil {
		return nil, err
	}

	var response *dns.Msg
	var err error

	for _, upstream := range handler.upstreams {
		response, _, err = handler.client.ExchangeContext(ctx, msg, upstream)
		if err != nil {
			continue
		}

		if response.Answer != nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	data, err := response.Pack()
	if err != nil {
		return nil, err
	}

	return data, nil
}
