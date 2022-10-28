package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
)

func (srv *Server) lambdaHandler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if srv.dnsHandler == nil {
		handler, err := NewDNSHandler()
		if err != nil {
			return events.APIGatewayV2HTTPResponse{}, err
		}

		srv.dnsHandler = handler
	}

	var query string
	switch req.RequestContext.HTTP.Method {
	case http.MethodGet:
		query = req.QueryStringParameters["dns"]

		if len(query) > 512 {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusRequestURITooLong,
			}, nil
		}

	case http.MethodPost:
		if req.Headers["content-type"] != "application/dns-message" {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusUnsupportedMediaType,
			}, nil
		}

		query = req.Body

		if len(query) > 512 {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusRequestEntityTooLarge,
			}, nil
		}

	default:
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusNotImplemented,
		}, nil

	}

	if len(query) < 12 {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	binary, err := base64.URLEncoding.DecodeString(query)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Body:       err.Error(),
		}, nil
	}

	body, err := srv.dnsHandler.Query(ctx, binary)
	if err != nil {
		log.Println("error handle:", err)

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode:      http.StatusOK,
		Body:            base64.RawURLEncoding.EncodeToString(body),
		IsBase64Encoded: true,
		Headers: map[string]string{
			"Content-Type":   "application/dns-message",
			"Content-Length": strconv.Itoa(len(body)),
		},
	}, nil
}
