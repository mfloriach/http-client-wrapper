package internal

import (
	"context"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/inhies/go-bytesize"
	"github.com/valyala/fasthttp"
)

type FastHttpClient interface {
	DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error
}

type HttpClient interface {
	Get(ctx context.Context, url string, cfg ...Configs) ([]byte, error)
}

type httpClient struct {
	client  FastHttpClient
	baseUrl string
	cfg     *configs
}

func NewHttpClient(client FastHttpClient, url string, cfg ...Configs) HttpClient {
	config := defaultConfigs()

	for _, fn := range cfg {
		fn(&config)
	}

	return &httpClient{client: client, baseUrl: url, cfg: &config}
}

func (h *httpClient) Get(ctx context.Context, path string, cfg ...Configs) ([]byte, error) {
	config := defaultConfigs()

	for _, fn := range cfg {
		fn(&config)
	}

	uri, err := url.ParseRequestURI(h.baseUrl + path)
	if err != nil {
		return []byte{}, err
	}

	var (
		method  = fasthttp.MethodGet
		req     = fasthttp.AcquireRequest()
		resp    = fasthttp.AcquireResponse()
		retries = 0
	)

	defer fasthttp.ReleaseResponse(resp)
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(uri.String())
	req.Header.SetMethod(method)
	req.Header.SetContentTypeBytes([]byte("application/json"))

	for k, v := range h.cfg.Headers {
		req.Header.Add(k, v)
	}

	for k, v := range config.Headers {
		req.Header.Add(k, v)
	}

	if config.Timeout != 0 {
		h.cfg.Timeout = config.Timeout
	}

	queries := uri.RawQuery
	if h.cfg.HideParams {
		queries = getQueries(uri.RawQuery)
	}

	for retries == 0 || (shouldRetry(err, resp) && retries < h.cfg.Retries) {
		time.Sleep(backoff(retries))

		if h.cfg.Throttle != nil && config.Throttle == nil {
			h.cfg.Throttle()
		} else if config.Throttle != nil {
			config.Throttle()
		}

		var start time.Time
		if h.cfg.CircuitBreaker != nil && config.CircuitBreaker == nil {
			h.cfg.CircuitBreaker(func() error {
				start = time.Now()
				return h.client.DoTimeout(req, resp, h.cfg.Timeout)
			})
		} else if config.CircuitBreaker != nil {
			start = time.Now()
			err = h.client.DoTimeout(req, resp, h.cfg.Timeout)
		} else {
			start = time.Now()
			err = h.client.DoTimeout(req, resp, h.cfg.Timeout)
		}

		slog.Info("Http call",
			"context", ctx,
			"retries", retries,
			"status_code", resp.StatusCode(),
			"duration", time.Since(start),
			"url", uri.Scheme+"://"+uri.Host+uri.Path,
			"queries", queries,
			"method", method,
			"size", bytesize.New(float64(len(resp.Body()))),
			"body", string(resp.Body())[:300]+"...",
		)

		retries++
	}

	return resp.Body(), nil
}

func backoff(retries int) time.Duration {
	return time.Duration(math.Pow(2, float64(retries))) * time.Second
}

func shouldRetry(err error, resp *fasthttp.Response) bool {
	if err != nil {
		return true
	}

	if resp.StatusCode() == http.StatusBadGateway ||
		resp.StatusCode() == http.StatusServiceUnavailable ||
		resp.StatusCode() == http.StatusGatewayTimeout {
		return true
	}

	return false
}

func getQueries(rawQueries string) string {
	queries := ""
	for _, q := range strings.Split(rawQueries, "&") {
		key := strings.Split(q, "=")[0]
		value := strings.Split(q, "=")[1]

		if len(value) > 4 {
			value = value[:3]
			value += "X"
			value += "X"
		} else {
			for i := len(value); i < 4; i++ {
				value += "X"
			}
		}

		queries += key + "=" + value + "&"
	}
	queries = queries[:len(queries)-1]

	return queries
}
