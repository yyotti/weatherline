package weatherline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const (
	lineNotifyAPIBase = "https://notify-api.line.me"
)

type notifyError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (e notifyError) Error() string {
	return fmt.Sprintf("%d: %s", e.Status, e.Message)
}

// LineNotify : Line Notify API client interface
type LineNotify interface {
	Send(string) error
}

type lineNotify struct {
	token string

	url        string
	httpClient *http.Client
}

// NewLineNotify : Create LineNotify instance
func NewLineNotify(token string) LineNotify {
	return &lineNotify{
		token: token,

		url:        lineNotifyAPIBase,
		httpClient: &http.Client{},
	}
}

// Send : NotifyClient.Send の実装
func (n *lineNotify) Send(msg string) error {
	values := url.Values{}
	values.Set("message", msg)

	u, err := url.Parse(n.url)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "api", "notify")

	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", n.token))

	res, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	response := notifyError{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	return response
}
