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

type Block struct {
	Number    int64
	Hash      string
	Owner     string
	BitmarkId string
}

type BlockResponse struct {
	Blocks []Block `json:"blocks"`
	Msg    string  `json:"message"`
}

type Transaction struct {
	Tx  json.RawMessage `json:"tx"`
	Msg string          `json:"message"`
}

type Bitmark struct {
	Bitmark json.RawMessage `json:"bitmark"`
	Msg     string          `json:"message"`
}

type Bitmarks struct {
	Bitmarks json.RawMessage `json:"bitmarks"`
	Msg      string          `json:"message"`
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

func (c Client) GetTx(txId string) ([]byte, error) {
	u := *c.u
	u.Path = fmt.Sprintf("/v1/txs/%s", txId)

	req, _ := http.NewRequest("GET", u.String(), nil)
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	var tx Transaction
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can not copy request body")
	}

	err = json.Unmarshal(buf.Bytes(), &tx)
	if err != nil {
		return nil, fmt.Errorf("error when parsing request body: %s", buf.Bytes())
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(tx.Msg)
	}

	return tx.Tx, nil
}

func (c Client) GetBitmarkByOwner(owner string, pending, provenance bool) ([]byte, error) {
	u := *c.u
	u.Path = "/v1/bitmarks"
	qs := url.Values{}
	if pending {
		qs.Set("pending", "true")
	}
	if provenance {
		qs.Set("provenance", "true")
	}
	if owner != "" {
		qs.Set("owner", owner)
	}
	u.RawQuery = qs.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	var bmk Bitmarks
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can not copy request body")
	}

	err = json.Unmarshal(buf.Bytes(), &bmk)
	if err != nil {
		return nil, fmt.Errorf("error when parsing request body: %s", buf.Bytes())
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(bmk.Msg)
	}
	return bmk.Bitmarks, nil
}

func (c Client) GetBitmark(bitmarkId string, pending, provenance bool) ([]byte, error) {
	u := *c.u
	u.Path = fmt.Sprintf("/v1/bitmarks/%s", bitmarkId)

	qs := url.Values{}
	if pending {
		qs.Add("pending", "true")
	}

	if provenance {
		qs.Add("provenance", "true")
	}
	u.RawQuery = qs.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	var bmk Bitmark
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can not copy request body")
	}

	err = json.Unmarshal(buf.Bytes(), &bmk)
	if err != nil {
		return nil, fmt.Errorf("error when parsing request body: %s", buf.Bytes())
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(bmk.Msg)
	}
	return bmk.Bitmark, nil
}

func (c Client) GetBlockNumber() (int64, error) {
	u := *c.u
	u.Path = fmt.Sprintf("/v1/blocks")

	req, _ := http.NewRequest("GET", u.String(), nil)
	resp, err := c.c.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var blockResp BlockResponse
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&blockResp)
	if err != nil {
		return 0, fmt.Errorf("error when parsing request body: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf(blockResp.Msg)
	}

	if len(blockResp.Blocks) == 0 {
		return 0, fmt.Errorf("invalid blocks return")
	}

	return blockResp.Blocks[0].Number, nil
}
