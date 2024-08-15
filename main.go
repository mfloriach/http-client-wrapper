package main

import (
	"context"
	"flag"
	"log/slog"
	"mfloriach90/interfaces/internal"
	"net/url"
	"os"
	"time"

	"github.com/sony/gobreaker/v2"
	"github.com/valyala/fasthttp"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	var endpoint string
	flag.StringVar(&endpoint, "url", "", "endpoint")
	flag.Parse()

	uri, err := url.ParseRequestURI(endpoint)
	if err != nil {
		logger.Error("could not parse", err.Error())
	}

	client := &fasthttp.Client{
		ReadTimeout:                   500 * time.Millisecond,
		WriteTimeout:                  500 * time.Millisecond,
		MaxIdleConnDuration:           24 * time.Hour,
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		MaxResponseBodySize:           200000000,
		MaxConnsPerHost:               500,
		// increase DNS cache time to an hour instead of default minute
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}

	c := internal.NewHttpClient(
		client,
		uri.Scheme+"://"+uri.Host,
		internal.AddHeaders(map[string]string{"Authentication": "Bearer abc1234567890"}),
		internal.AddRetries(5),
		internal.HideParams(),
		internal.AddThrottle(3),
		internal.AddCircuitBreaker(gobreaker.Settings{
			Name: "HTTP GET",
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				return counts.Requests >= 3 && failureRatio >= 0.6
			},
		}),
	)

	for i := 0; i < 10; i++ {
		ctx := context.Background()
		if _, err := c.Get(ctx, uri.Path+"?"+uri.RawQuery, internal.AddTimeout(5*time.Second)); err != nil {
			logger.ErrorContext(ctx, "http request failed", "error", err)
		}
	}
}
