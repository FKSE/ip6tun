package ip6tun

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type TunnelResponse struct {
	Id         uint32     `json:"id"`
	Name       string     `json:"name"`
	LocalPort  uint16     `json:"local_port"`
	RemoteHost string     `json:"remote_host"`
	RemotePort uint16     `json:"remote_port"`
	MessageLog []string   `json:"message_log"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

// reqTunnel is used for create & update
type TunnelRequest struct {
	Name       string `json:"name"`
	LocalPort  uint16 `json:"local_port"`
	RemotePort uint16 `json:"remote_port"`
}

// Validate checks if given ports are in range and if the name is not empty.
func (t *TunnelRequest) Validate() error {
	if t.Name == "" {
		return errors.New("The Name may not be empty.")
	}
	if t.LocalPort < 1 || t.LocalPort > 65535 {
		return errors.New("The LocalPort must be between 1.")
	}
	if t.RemotePort < 1 || t.RemotePort > 65535 {
		return errors.New("The RemotePort must be between 1.")
	}
	return nil
}

// Client is a simple REST-Client to access a ip6tun server
type Client struct {
	address string
	apiKey  string
	http    *http.Client
}

// NewClient instantiates a new client which is able to communicate with the server at host:port
func NewClient(host string, port int, apiKey string) *Client {
	return &Client{
		address: fmt.Sprintf("https://%s:%d", host, port),
		apiKey:  apiKey,
		http:    http.DefaultClient,
	}
}

// List returns a list of all active tunnels
func (c *Client) List() (list []TunnelResponse, err error) {
	r, err := http.Get(c.address)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	// decode response
	err = handleResponse(r, http.StatusOK, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// Create a new tunnel on the server
func (c *Client) Create(name string, serverPort, clientPort uint16) (*TunnelResponse, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(TunnelRequest{name, serverPort, clientPort})
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest("POST", "", &buf)
	if err != nil {
		return nil, err
	}
	// make the request
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var tun TunnelResponse
	err = handleResponse(res, http.StatusCreated, &tun)
	if err != nil {
		return nil, err
	}
	return &tun, nil
}

// Update the tunnel on the server by its id
func (c *Client) Update(id uint32, name string, serverPort, clientPort uint16) (*TunnelResponse, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(TunnelRequest{name, serverPort, clientPort})
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest("PUT", fmt.Sprintf("%d", id), &buf)
	if err != nil {
		return nil, err
	}
	// make the request
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var tun TunnelResponse
	err = handleResponse(res, http.StatusOK, &tun)
	if err != nil {
		return nil, err
	}
	return &tun, nil
}

func (c *Client) Delete() (*TunnelResponse, error) {
	return nil, nil
}

// newRequest prepares a request
func (c *Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	url := c.address + "/" + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set(HTTPAuthHeader, c.apiKey)

	return req, nil
}

func handleResponse(r *http.Response, expectedStatus int, v interface{}) error {
	// handle errors
	if r.StatusCode != expectedStatus {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		return errors.New(strings.TrimSpace(string(b)))
	}
	// decode json payload
	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		return err
	}
	return nil
}
