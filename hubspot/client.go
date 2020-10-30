package hubspot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
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

// AssociationErrorResponse handles the HubSpot error when associating two objects
type AssociationErrorResponse struct {
	Status     string
	StatusCode int
	Message    string
	NumErrors  int
	Errors     []ErrorResponse
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

// BuildAssociationURL returns the URL for HubSpot object to object associations
func BuildAssociationURL(c *Client, from string, to string) (string, error) {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if len(from) == 0 || len(to) == 0 {
		return "", fmt.Errorf("from and to arguments require a value")
	}
	return fmt.Sprintf("%s/crm/%s/associations/%s/%s/batch/create?hapikey=%s", c.APIBaseURL, c.APIVersion, from, to, c.APIKey), nil
}

// CreateAssociation relates two objects to each other in HubSpot
func (c *Client) CreateAssociation(association *AssociationInput, from string, to string) (*AssociationResults, AssociationErrorResponse) {

	requestBody, err := json.Marshal(association)
	if err != nil {
		return nil, AssociationErrorResponse{Status: "error", Message: "invalid association input"}
	}

	requestURL, err := BuildAssociationURL(c, from, to)
	if err != nil {
		return nil, AssociationErrorResponse{Status: "error", Message: fmt.Sprintf("unable to build url, err: %v", err)}
	}

	r, err := c.request(
		requestURL,
		http.MethodPost,
		requestBody)

	if err != nil {
		return nil,
			AssociationErrorResponse{Status: "error", Message: fmt.Sprintf("unable to execute request, err: %v", err)}
	}

	if r.StatusCode != http.StatusCreated {
		var errorResponse AssociationErrorResponse
		err := json.Unmarshal(r.Body, &errorResponse)
		errorResponse.StatusCode = r.StatusCode

		if err != nil {
			errorResponse.Status = "error"
			errorResponse.Message = fmt.Sprintf("Unable to unmarshal HubSpot association error response, err: %v", err)
		}
		return nil, errorResponse
	}

	var associationResult AssociationResults
	if err := json.Unmarshal(r.Body, &associationResult); err != nil {
		msg := fmt.Sprintf("could not unmarshal HubSpot response, err: %v", err)
		return nil, AssociationErrorResponse{Status: "error", Message: msg}
	}

	return &associationResult, AssociationErrorResponse{}
}

// CreateContact creates a new Contact in HubSpot
func (c *Client) CreateContact(contactInput *ContactInput) (*ContactOutput, ErrorResponse) {
	requestBody, err := json.Marshal(contactInput)
	if err != nil {
		return nil, ErrorResponse{Status: "error", Message: "invalid contact input"}
	}
	r, err := c.request(
		fmt.Sprintf("%s/crm/%s/objects/contacts/?hapikey=%s", c.APIBaseURL, c.APIVersion, c.APIKey),
		http.MethodPost,
		requestBody)

	if err != nil {
		return nil,
			ErrorResponse{Status: "error", Message: fmt.Sprintf("unable to execute request, err: %v", err)}
	}

	if r.StatusCode != http.StatusCreated {
		var errorResponse ErrorResponse
		err := json.Unmarshal(r.Body, &errorResponse)
		errorResponse.StatusCode = r.StatusCode

		if err != nil {
			errorResponse.Status = "error"
			errorResponse.Message = fmt.Sprintf("Unable to unmarshal HubSpot create account error response, err: %v", err)
		}
		return nil, errorResponse
	}

	var contactOutput ContactOutput
	if err := json.Unmarshal(r.Body, &contactOutput); err != nil {
		msg := fmt.Sprintf("could not unmarshal HubSpot response, err: %v", err)
		return nil, ErrorResponse{Status: "error", Message: msg}
	}

	return &contactOutput, ErrorResponse{}
}

// UpdateContact updates a Contact in HubSpot
func (c *Client) UpdateContact(contactID string, contactInput *ContactInput) (*ContactOutput, ErrorResponse) {
	requestBody, err := json.Marshal(contactInput)
	if err != nil {
		return nil, ErrorResponse{Status: "error", Message: "invalid contact input"}
	}

	apiURL := fmt.Sprintf("%s/crm/%s/objects/contacts/%s?hapikey=%s", c.APIBaseURL, c.APIVersion, contactID, c.APIKey)
	r, err := c.request(apiURL, http.MethodPatch, requestBody)

	if err != nil {
		return nil,
			ErrorResponse{Status: "error", Message: fmt.Sprintf("unable to execute request, err: %v", err)}
	}

	if r.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		err := json.Unmarshal(r.Body, &errorResponse)
		errorResponse.StatusCode = r.StatusCode
		if err != nil {
			errorResponse.Status = "error"
			errorResponse.Message = fmt.Sprintf("unable to unmarshal HubSpot update account error response, err: %v", err)
		}

		return nil, errorResponse
	}

	var contactOutput ContactOutput
	if err := json.Unmarshal(r.Body, &contactOutput); err != nil {
		msg := fmt.Sprintf("could not unmarshal HubSpot response, err: %v", err)
		return nil, ErrorResponse{Status: "error", Message: msg}
	}

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
		return nil, errors.New("request execution failed")
	}

	req.Header.Add("Content-Type", "application/json")
	r, err := c.HTTPClient.Do(req)
	if err != nil {
		return &response, errors.New("request execution failed")
	}

	defer r.Body.Close()

	// prepare response
	response.StatusCode = r.StatusCode
	if err := json.NewDecoder(r.Body).Decode(&response.Body); err != nil {
		return &response, errors.New("ERROR: could not decode response")
	}

	return &response, nil
}
