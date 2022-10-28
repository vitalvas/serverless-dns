package main

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
)

func (srv *Server) lambdaHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if srv.dnsHandler == nil {
		srv.dnsHandler = NewDNSHandler(&DNSConfig{
			timeout: 4,
			upstreams: []string{
				// Cloudflare
				"1.1.1.1:53",
				"1.0.0.1:53",

				// Google
				"8.8.8.8:53",
				"8.8.4.4:53",
			},
		})
	}

	query, err := parseQuery(req)
	if err != nil || len(query) == 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       err.Error(),
		}, nil
	}

	body, err := srv.dnsHandler.Query(ctx, query)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode:      http.StatusOK,
		Body:            base64.StdEncoding.EncodeToString(body),
		IsBase64Encoded: true,
		Headers: map[string]string{
			"Content-Type":   "application/dns-message",
			"Content-Length": strconv.Itoa(len(body)),
		},
	}, nil
}

func parseQuery(req events.APIGatewayProxyRequest) (string, error) {
	switch req.HTTPMethod {
	case http.MethodGet:
		return req.QueryStringParameters["dns"], nil

	case http.MethodPost:
		return req.Body, nil

	}

	return "", errors.New("Invalid DNS request")
}
