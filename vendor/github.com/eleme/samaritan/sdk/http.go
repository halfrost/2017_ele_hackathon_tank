package sdk

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type baseHTTPClient struct {
	c         *http.Client
	userAgent string
}

func (c *baseHTTPClient) setUserAgent(ua string) {
	c.userAgent = ua
}

// DoSimple requests the given url by given method.
// DoSimple drops the response.
func (c *baseHTTPClient) doSimple(method, url string) error {
	res, err := c.do(method, url)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}
	return err
}

// Do requests the given url by given method.
// Do returns error if request fail or the response's status code is not http.StatusOK.
func (c *baseHTTPClient) do(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return nil, fmt.Errorf("%v %s %s: %s", res.StatusCode, method, url, body)
	}

	return res, nil
}

func newBaseClient() *baseHTTPClient {
	return &baseHTTPClient{
		c: &http.Client{},
	}
}
