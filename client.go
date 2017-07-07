package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Bitmark struct {
	Bitmark json.RawMessage `json:"bitmark"`
	Message string          `json:"message"`
}

type Client struct {
	u *url.URL
	c *http.Client
}

func New(serverUri string) (*Client, error) {
	c := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	u, err := url.Parse(serverUri)
	if err != nil {
		return nil, err
	}

	return &Client{
		u: u,
		c: c,
	}, nil
}

func (c Client) GetBitmark(txId string) ([]byte, error) {
	u := *c.u
	u.Path = fmt.Sprintf("/registry/v1/bitmarks/%s", txId)

	req, _ := http.NewRequest("GET", u.String(), nil)
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	var bitmark Bitmark
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can not copy request body")
	}

	err = json.Unmarshal(buf.Bytes(), &bitmark)
	if err != nil {
		return nil, fmt.Errorf("error when parsing request body: %s", buf.Bytes())
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(bitmark.Message)
	}

	return bitmark.Bitmark, nil
}