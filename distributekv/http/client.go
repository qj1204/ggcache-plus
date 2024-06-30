package http

import (
	"fmt"

	"io"
	"net/http"
	"net/url"
)

type httpClient struct {
	baseURL string // the address of the remote node to be accessed. such as http://example.com/ggroupcache/
}

// Fetch responsible for querying the value of key from the group cache of the specified node through http request
func (h *httpClient) Fetch(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	fmt.Println(u)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %v", err)
	}

	return bytes, nil
}
