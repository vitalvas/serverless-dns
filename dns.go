package main

import (
	"context"
	"encoding/base64"
	"math/rand"
	"net"
	"time"

	"github.com/miekg/dns"
)

type DNSConfig struct {
	timeout   int
	upstreams []string
}

type DNSHandler struct {
	upstreams []string
	client    *dns.Client
}

func NewDNSHandler(config *DNSConfig) *DNSHandler {
	timeout := time.Duration(config.timeout) * time.Second

	handler := &DNSHandler{
		upstreams: config.upstreams,
		client: &dns.Client{
			Net:     "udp",
			Timeout: timeout,
			Dialer: &net.Dialer{
				Timeout:   timeout,
				LocalAddr: nil,
			},
		},
	}

	return handler
}

func (handler *DNSHandler) randomUpstream() string {
	return handler.upstreams[rand.Intn(len(handler.upstreams))]
}

func (handler *DNSHandler) Query(ctx context.Context, query string) ([]byte, error) {
	binary, err := base64.RawURLEncoding.DecodeString(query)
	if err != nil {
		return nil, err
	}

	msg := new(dns.Msg)
	msg.Unpack(binary)

	upstream := handler.randomUpstream()

	response, _, err := handler.client.ExchangeContext(ctx, msg, upstream)
	if err != nil {
		return nil, err
	}

	data, err := response.Pack()
	if err != nil {
		return nil, err
	}

	return data, nil
}
