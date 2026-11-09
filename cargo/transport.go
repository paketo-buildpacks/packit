package cargo

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/avast/retry-go/v5"
)

type Transport struct {
	logger     io.Writer
	retryCount uint
	delay      time.Duration
}

type TransportOption func(Transport) Transport

func NewTransport(options ...TransportOption) Transport {
	t := Transport{logger: io.Discard, retryCount: 5, delay: 100 * time.Millisecond}
	for _, option := range options {
		t = option(t)
	}
	return t
}

func WithLogger(w io.Writer) TransportOption {
	return func(t Transport) Transport {
		t.logger = w
		return t
	}
}

func WithRetryCount(count uint) TransportOption {
	return func(t Transport) Transport {
		t.retryCount = count
		return t
	}
}

func WithRetryDelay(delay time.Duration) TransportOption {
	return func(t Transport) Transport {
		t.delay = delay
		return t
	}
}

func (t Transport) Drop(root, uri string) (io.ReadCloser, error) {
	if strings.HasPrefix(uri, "file://") {
		file, err := os.Open(filepath.Join(root, strings.TrimPrefix(uri, "file://")))
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %s", err)
		}

		return file, nil
	}

	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request uri: %s", err)
	}

	var response *http.Response
	attempt := 0
	err = retry.New(
		retry.Attempts(t.retryCount),
		retry.Delay(t.delay),
	).Do(
		func() error {
			var err error
			if attempt > 0 {
				_, _ = fmt.Fprintf(t.logger, "Retrying request to %q (attempt %d/%d)\n", uri, attempt+1, t.retryCount)
			}
			attempt++
			response, err = http.DefaultClient.Do(request)
			if err != nil {
				return fmt.Errorf("failed to make request: %s", err)
			}
			if response.StatusCode >= 400 {
				_ = response.Body.Close()
				return fmt.Errorf("unexpected status code %d while fetching %q", response.StatusCode, uri)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return response.Body, nil
}
