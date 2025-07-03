package aoscxgo

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// fetchLatestAPIVersion fetches the latest API version from the switch
func fetchLatestAPIVersion(transport *http.Transport, hostname string) (string, error) {
	url := fmt.Sprintf("https://%s/rest", hostname)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create API version request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Close = false

	res, err := transport.RoundTrip(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch API versions: %w", err)
	}
	if res == nil {
		return "", fmt.Errorf("received nil response when fetching API versions")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch API versions, status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read API version response: %w", err)
	}

	// Parse JSON response and extract latest version
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", fmt.Errorf("failed to parse API version response: %w", err)
	}

	// Get the latest version
	latest, exists := apiResponse["latest"]
	if !exists {
		return "", fmt.Errorf("latest version not found in API response")
	}

	latestMap, ok := latest.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("latest version has invalid format")
	}

	version, exists := latestMap["version"]
	if !exists {
		return "", fmt.Errorf("version not found in latest API response")
	}

	versionStr, ok := version.(string)
	if !ok || versionStr == "" {
		return "", fmt.Errorf("version is not a valid string")
	}

	log.Printf("Detected latest API version: %s", versionStr)
	return versionStr, nil
}

// Connect creates connection to given Client object.
func Connect(c *Client) (*Client, error) {
	var err error

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !c.VerifyCertificate},
	}

	if c.Transport == nil {
		c.Transport = tr
	}

	// Fetch the latest API version from the switch
	latestVersion, err := fetchLatestAPIVersion(c.Transport, c.Hostname)
	if err != nil {
		log.Printf("Warning: Could not fetch latest API version, using fallback: %v", err)
		// Fall back to user-specified version or default
		if c.Version == "" {
			c.Version = "v10.09" // Default fallback
		}
	} else {
		// Use the latest version from the switch
		c.Version = latestVersion
		log.Printf("Using API version: %s", c.Version)
	}

	// Validate version format (should start with 'v')
	if c.Version != "" && c.Version[0] != 'v' {
		c.Version = "v" + c.Version
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
