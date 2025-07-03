package aoscxgo

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type Client struct {
	// Connection properties.
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	Password string `json:"password"`
	Version  string `json:"version"`
	// Generated after Connect
	Cookie *http.Cookie `json:"cookie"`
	Csrf   string       `json:"Csrf"`
	// HTTP transport options.  Note that the VerifyCertificate setting is
	// only used if you do not specify a HTTP transport yourself.
	VerifyCertificate bool            `json:"verify_certificate"`
	Transport         *http.Transport `json:"-"`
}

// Connect creates connection to given Client object.
func Connect(c *Client) (*Client, error) {
	var err error

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !c.VerifyCertificate},
	}

	if c.Version == "" || (c.Version != "v10.09" && c.Version == "v10.10") {
		c.Version = "v10.09"
	}

	if c.Transport == nil {
		c.Transport = tr
	}

	cookie, csrf, err := login(c.Transport, c.Hostname, c.Version, c.Username, c.Password)

	if err != nil {
		return nil, err
	}
	c.Cookie = cookie
	c.Csrf = csrf
	return c, err
}

// Logout calls the logout endpoint to clear the session.
func (c *Client) Logout() error {
	if c == nil {
		return errors.New("nil value to Logout")
	}
	url := fmt.Sprintf("https://%s/rest/%s/logout", c.Hostname, c.Version)
	resp := logout(c.Transport, c.Cookie, c.Csrf, url)
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}

// login performs POST to create a cookie for authentication to the given IP with the provided credentials.
func login(http_transport *http.Transport, ip string, rest_version string, username string, password string) (*http.Cookie, string, error) {
	url := fmt.Sprintf("https://%s/rest/%s/login?username=%s&password=%s", ip, rest_version, username, password)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("accept", "*/*")
	req.Header.Set("x-use-csrf-token", "true")
	req.Close = false

	res, err := http_transport.RoundTrip(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to connect to switch: %w", err)
	}
	if res == nil {
		return nil, "", fmt.Errorf("received nil response from switch")
	}
	if res.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("login failed with status: %s", res.Status)
	}

	log.Printf("Login Successful")

	// Check if CSRF token exists
	csrfTokens := res.Header["X-Csrf-Token"]
	if len(csrfTokens) == 0 {
		return nil, "", fmt.Errorf("no CSRF token received from switch")
	}
	csrf := csrfTokens[0]

	// Check if cookies exist
	cookies := res.Cookies()
	if len(cookies) == 0 {
		return nil, "", fmt.Errorf("no cookies received from switch")
	}
	cookie := cookies[0]

	return cookie, csrf, nil
}

// logout performs POST to logout using a cookie from the given URL.
func logout(http_transport *http.Transport, cookie *http.Cookie, csrf string, url string) *http.Response {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Printf("Failed to create logout request: %v", err)
		return nil
	}
	req.Header.Set("accept", "*/*")
	req.Header.Set("x-csrf-token", csrf)
	req.Close = false

	req.AddCookie(cookie)
	res, err := http_transport.RoundTrip(req)
	if err != nil {
		log.Printf("Failed to logout from switch: %v", err)
		return nil
	}
	if res == nil {
		log.Printf("Received nil response during logout")
		return nil
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("Logout failed with status: %s", res.Status)
		return res
	}

	log.Printf("Logout Successful")

	return res
}
