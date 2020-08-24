package hubspot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

// HubSpot defaults
const (
	DefaultAPIBaseURL = "https://api.hubapi.com"
	DefaultAPIVersion = "v3"
)

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client allows you to create a new HubSpot client
type Client struct {
	APIBaseURL string
	APIKey     string
	APIVersion string
	HTTPClient HTTPClient
}

// ErrorResponse handles the error structure returned by HubSpot API
type ErrorResponse struct {
	Category      string
	CorrelationID string
	Links         map[string]string
	Message       string
	Status        string
	StatusCode    int
}

// Response handles a response by the request method
type Response struct {
	Body       json.RawMessage
	StatusCode int
}

// NewClient creates a new HubSpot Client with corresponding defaults
func NewClient(apiKey string) *Client {
	c := &Client{}
	c.APIKey = apiKey
	c.APIBaseURL = DefaultAPIBaseURL
	c.APIVersion = DefaultAPIVersion

	// Instantiate gzip client with a 5 second timeout on waiting for the
	// remote server to accept the connection and a 30 second timeout
	// for no activity over the connection
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		MaxIdleConns:       2,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	c.HTTPClient = &http.Client{
		Transport: transport,
	}
	return c
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf(e.Message)
}

// CreateContact creates a new Contact in HubSpot
func (c *Client) CreateContact(body Contact) (*Contact, ErrorResponse) {
	var contact Contact
	log.Printf("INFO: attempting to create HubSpot Contact")

	requestBody, err := json.Marshal(body)
	if err != nil {
		log.Printf("ERROR: could not marshal the provided contact body, err: %v", err)
		return &contact, ErrorResponse{Status: "error", Message: "invalid contact body"}
	}
	r, err := c.request(
		fmt.Sprintf("%s/crm/%s/objects/contacts/?hapikey=%s", c.APIBaseURL, c.APIVersion, c.APIKey),
		http.MethodPost,
		requestBody)

	if err != nil {
		log.Printf("ERROR: unable to create hubspot contact")
		return &contact,
			ErrorResponse{Status: "error", Message: fmt.Sprintf("unable to execute request, err: %v", err)}
	}

	if r.StatusCode != http.StatusCreated {
		var errorResponse ErrorResponse
		err := json.Unmarshal(r.Body, &errorResponse)
		msg := "ERROR: unable to create HubSpot account. "
		if err != nil {
			log.Printf("%sUnable to unmarshall error response.", msg)
		} else {
			log.Printf("%sGot error: %v.", msg, errorResponse.Message)
		}
		errorResponse.StatusCode = r.StatusCode
		return &contact, errorResponse
	}

	if err := json.Unmarshal(r.Body, &contact); err != nil {
		msg := fmt.Sprintf("could not unmarshal HubSpot response, err: %v", err)
		log.Printf("ERROR: %s", msg)
		return &contact, ErrorResponse{Status: "error", Message: msg}
	}

	log.Printf("INFO: HubSpot contact created successfully. Contact ID: %s", contact.ID)
	return &contact, ErrorResponse{}
}

// request executes a HTTP request and returns the response
func (c *Client) request(
	url string,
	method string,
	requestBody []byte) (*Response, error) {

	var response Response

	// Timeout the entire request after 30 seconds if the server accepts the connection
	// but never responds
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("ERROR: unable to create a HubSpot request, got error = %v", err)
		return &response, errors.New("request execution failed")
	}

	req.Header.Add("Content-Type", "application/json")
	r, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("ERROR: unable to complete request, got error = %v", err)
		return &response, errors.New("request execution failed")
	}

	log.Printf("INFO: request successful, got response status: %v", r.StatusCode)

	// prepare response
	response.StatusCode = r.StatusCode
	if err := json.NewDecoder(r.Body).Decode(&response.Body); err != nil {
		log.Printf("ERROR: could not decode HubSpot response, err: %v", err)
		return &response, errors.New("ERROR: could not decode response")
	}

	defer r.Body.Close()

	return &response, nil
}
