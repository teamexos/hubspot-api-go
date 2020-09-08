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
func (c *Client) CreateContact(contactInput *ContactInput) (*ContactOutput, ErrorResponse) {
	log.Printf("INFO: attempting to create HubSpot Contact")

	requestBody, err := json.Marshal(contactInput)
	if err != nil {
		log.Printf("ERROR: could not marshal the provided contact body, err: %v", err)
		return nil, ErrorResponse{Status: "error", Message: "invalid contact input"}
	}
	r, err := c.request(
		fmt.Sprintf("%s/crm/%s/objects/contacts/?hapikey=%s", c.APIBaseURL, c.APIVersion, c.APIKey),
		http.MethodPost,
		requestBody)

	if err != nil {
		log.Printf("ERROR: unable to create hubSpot contact")
		return nil,
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
		return nil, errorResponse
	}

	var contactOutput ContactOutput
	if err := json.Unmarshal(r.Body, &contactOutput); err != nil {
		msg := fmt.Sprintf("could not unmarshal HubSpot response, err: %v", err)
		log.Printf("ERROR: %s", msg)
		return nil, ErrorResponse{Status: "error", Message: msg}
	}

	log.Printf("INFO: HubSpot contact created successfully. Contact ID: %s", contactOutput.ID)
	return &contactOutput, ErrorResponse{}
}

// UpdateContact updates a Contact in HubSpot
func (c *Client) UpdateContact(contactID string, contactInput *ContactInput) (*ContactOutput, ErrorResponse) {
	log.Printf("INFO: attempting to update HubSpot Contact")

	//contactInput.Properties.
	requestBody, err := json.Marshal(contactInput)
	if err != nil {
		log.Printf("ERROR: could not marshal the provided contact body, err: %v", err)
		return nil, ErrorResponse{Status: "error", Message: "invalid contact input"}
	}
	r, err := c.request(
		fmt.Sprintf("%s/crm/%s/objects/contacts/%s?hapikey=%s", c.APIBaseURL, c.APIVersion, contactID, c.APIKey),
		http.MethodPatch,
		requestBody)

	if err != nil {
		log.Printf("ERROR: unable to update hubSpot contact")
		return nil,
			ErrorResponse{Status: "error", Message: fmt.Sprintf("unable to execute request, err: %v", err)}
	}

	if r.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		err := json.Unmarshal(r.Body, &errorResponse)
		msg := "ERROR: unable to update HubSpot account. "
		if err != nil {
			log.Printf("%sUnable to unmarshall error response.", msg)
		} else {
			log.Printf("%sGot error: %v.", msg, errorResponse.Message)
		}
		errorResponse.StatusCode = r.StatusCode
		return nil, errorResponse
	}

	var contactOutput ContactOutput
	if err := json.Unmarshal(r.Body, &contactOutput); err != nil {
		msg := fmt.Sprintf("could not unmarshal HubSpot response, err: %v", err)
		log.Printf("ERROR: %s", msg)
		return nil, ErrorResponse{Status: "error", Message: msg}
	}

	log.Printf("INFO: HubSpot contact updated successfully. Contact ID: %s", contactOutput.ID)
	return &contactOutput, ErrorResponse{}
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
		return nil, errors.New("request execution failed")
	}

	req.Header.Add("Content-Type", "application/json")
	r, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("ERROR: unable to complete request, got error = %v", err)
		return &response, errors.New("request execution failed")
	}

	defer r.Body.Close()

	log.Printf("INFO: request successful, got response status: %v", r.StatusCode)

	// prepare response
	response.StatusCode = r.StatusCode
	if err := json.NewDecoder(r.Body).Decode(&response.Body); err != nil {
		log.Printf("ERROR: could not decode HubSpot response, err: %v", err)
		return &response, errors.New("ERROR: could not decode response")
	}

	return &response, nil
}
