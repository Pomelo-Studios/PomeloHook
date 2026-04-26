package forward

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type ForwardResult struct {
	EventID    string
	StatusCode int
	Body       string
	MS         int64
}

type Forwarder struct {
	targetBaseURL string
	client        *http.Client
}

func New(targetBaseURL string) *Forwarder {
	return &Forwarder{
		targetBaseURL: targetBaseURL,
		client:        &http.Client{Timeout: 10 * time.Second},
	}
}

type incomingPayload struct {
	EventID string `json:"event_id"`
	Method  string `json:"method"`
	Path    string `json:"path"`
	Headers string `json:"headers"`
	Body    string `json:"body"`
}

func (f *Forwarder) Forward(raw []byte) (*ForwardResult, error) {
	var p incomingPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(p.Method, f.targetBaseURL+p.Path, bytes.NewBufferString(p.Body))
	if err != nil {
		return nil, err
	}
	var headers map[string][]string
	if err := json.Unmarshal([]byte(p.Headers), &headers); err == nil {
		for k, vals := range headers {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}
	}

	start := time.Now()
	resp, err := f.client.Do(req)
	ms := time.Since(start).Milliseconds()
	if err != nil {
		return &ForwardResult{EventID: p.EventID, StatusCode: 0, MS: ms}, err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	return &ForwardResult{
		EventID:    p.EventID,
		StatusCode: resp.StatusCode,
		Body:       string(bodyBytes),
		MS:         ms,
	}, nil
}
