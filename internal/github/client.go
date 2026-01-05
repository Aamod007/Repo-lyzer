package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Client struct {
	http  *http.Client
	token string
}

type User struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

func NewClient() *Client {
	return &Client{
		http:  &http.Client{},
		token: os.Getenv("GITHUB_TOKEN"),
	}
}

// get performs a GET request to the GitHub API and decodes the JSON response.
// It handles authentication using the client's token if available.
// Parameters:
//   - url: The GitHub API endpoint URL
//   - target: Pointer to struct where the JSON response will be decoded
//
// Returns an error if the request fails or the response cannot be decoded.
func (c *Client) get(url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"GitHub API error: %s (tip: set GITHUB_TOKEN env variable)",
			resp.Status,
		)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (c *Client) GetUser() (*User, error) {
	var u User
	err := c.get("https://api.github.com/user", &u)
	return &u, err
}
