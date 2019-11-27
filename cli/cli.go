package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type ScreepsCLI struct {
	client      *resty.Client
	WelcomeText string
}

func NewScreepsCLI(host string, port int16) *ScreepsCLI {
	client := resty.New()
	client.SetHostURL(fmt.Sprintf("http://%s:%d", host, port))
	s := &ScreepsCLI{
		client: client,
	}
	resp, err := client.R().Get("/greeting")
	if err == nil {
		s.WelcomeText = resp.String()
	}
	return s
}

func (s *ScreepsCLI) Start() error {
	return nil
}

func (s *ScreepsCLI) Stop() {

}

func (s *ScreepsCLI) Command(cmd string) string {
	if len(cmd) == 0 {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(cmd).
		Post("/cli")
	if err != nil {
		return err.Error()
	}
	return resp.String()
}
