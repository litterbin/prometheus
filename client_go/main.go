package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "embed"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

var (
	//go:embed token
	_token string

	//go:embed address
	_address string

	token   = strings.TrimSpace(_token)
	address = strings.TrimSpace(_address)
)

type authRoundTripper struct {
	token string
	rt    http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface.
func (a authRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	var bearer = "Bearer " + a.token

	// The specification of http.RoundTripper says that it shouldn't mutate
	// the request so make a copy of req.Header since this is all that is
	// modified.
	r2 := new(http.Request)
	*r2 = *r
	r2.Header.Add("Authorization", bearer)

	r = r2
	return a.rt.RoundTrip(r)
}

func main() {
	client, err := api.NewClient(api.Config{
		Address: address,
		RoundTripper: authRoundTripper{
			rt:    api.DefaultRoundTripper,
			token: token,
		},
	})
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}

	api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := api.Query(ctx, "up{job=\"prometheus-k8s\"}", time.Now(), v1.WithTimeout(1*time.Second))
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("warnings: %v\n", warnings)
		os.Exit(1)
	}

	fmt.Printf("Result:\n%v\n", result)
}
