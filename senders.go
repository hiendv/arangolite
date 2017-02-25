package arangolite

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type sender interface {
	Send(cli *http.Client, req *http.Request) (*response, error)
}

type basicSender struct{}

func (s *basicSender) Send(cli *http.Client, req *http.Request) (*response, error) {
	res, err := cli.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "the database HTTP request failed")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		res.Body.Close()
		return nil, errors.Errorf("the database HTTP request failed, status code %d", res.StatusCode)
	}

	raw, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not read the database response")
	}
	parsed := parsedResponse{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, errors.Wrap(err, "database response decoding failed")
	}

	return &response{raw: raw, parsed: parsed}, nil
}

type parsedResponse struct {
	Error        bool            `json:"error"`
	ErrorMessage string          `json:"errorMessage"`
	Result       json.RawMessage `json:"result"`
	Cached       bool            `json:"cached"`
	HasMore      bool            `json:"hasMore"`
	ID           string          `json:"id"`
}

type response struct {
	raw    json.RawMessage
	parsed parsedResponse
}

func (r *response) Raw() json.RawMessage {
	return r.raw
}

func (r *response) RawResult() json.RawMessage {
	return r.parsed.Result
}

func (r *response) HasMore() bool {
	return r.parsed.HasMore
}

func (r *response) Cursor() string {
	return r.parsed.ID
}

func (r *response) Unmarshal(v interface{}) error {
	return json.Unmarshal(r.raw, v)
}