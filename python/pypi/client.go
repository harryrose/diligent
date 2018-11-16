package pypi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ProjectInfo struct {
	License string `json:"license"`
}

type ProjectMetadata struct {
	Info ProjectInfo `json:"info"`
}

type Client struct {
	BaseURL string
}

func (c *Client) ProjectMetadata(projectName string, projectVersion string) (*ProjectMetadata, error) {
	client := &http.Client{}

	parts := []string {
		c.BaseURL,
		"pypi",
		url.PathEscape(projectName),
	}

	if projectVersion != "" {
		parts = append(parts, url.PathEscape(projectVersion))
	}

	url := fmt.Sprintf("%s/json",strings.Join(parts, "/"))

	req, err := http.NewRequest(http.MethodGet, url, nil )

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var out ProjectMetadata
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&out)

	if err != nil {
		return nil, err
	}
	return &out, nil
}

