package dto

import (
	"github.com/blent/beagle/src/core/notification"
	"github.com/pkg/errors"
	"strings"
)

type Endpoint struct {
	*notification.Endpoint
}

func ToEndpoint(epDto *Endpoint) (*notification.Endpoint, error) {
	epDto.Name = strings.TrimSpace(epDto.Name)

	if epDto.Name == "" {
		return nil, errors.New("missed endpoint name")
	}

	epDto.Url = strings.TrimSpace(epDto.Url)

	if epDto.Url == "" {
		return nil, errors.New("missed endpoint url")
	}

	epDto.Method = strings.TrimSpace(epDto.Method)

	if epDto.Method == "" {
		return nil, errors.New("missed endpoint method")
	}

	return &notification.Endpoint{
		Id:      epDto.Id,
		Name:    epDto.Name,
		Url:     epDto.Url,
		Method:  epDto.Method,
		Headers: epDto.Headers,
	}, nil
}

func FromEndpoint(ep *notification.Endpoint) (*Endpoint, error) {
	return &Endpoint{
		Endpoint: ep,
	}, nil
}
