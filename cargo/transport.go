package cargo

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Transport struct{
	header http.Header
}

func NewTransport() Transport {
	return Transport{}
}

func (t Transport) Drop(root, uri string, header http.Header) (io.ReadCloser, error) {
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

	if header != nil {
		request.Header = header
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %s", err)
	}

	return response.Body, nil
}

