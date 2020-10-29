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

//CreateAssocation relates two objects to each other in HubSpot
func (c *Client) CreateAssociation(association *AssociationInput, fromObject string, toObject string) (*AssociationResults, AssociationErrorResponse) {
	log.Printf("INFO: attempting to create HubSpot object association")

	requestBody, err := json.Marshal(association)
	if err != nil {
		log.Printf("ERROR: could not marshal the provided object association body, err: %v", err)
		return nil, AssociationErrorResponse{Status: "error", Message: "invalid association input"}
	}

	r, err := c.request(
		fmt.Sprintf("%s/crm/%s/associations/%s/%s/batch/create?hapikey=%s", c.APIBaseURL, c.APIVersion, fromObject, toObject, c.APIKey),
		http.MethodPost,
		requestBody)

	if err != nil {
		log.Printf("ERROR: unable to create HubSpot association")
		return nil,
			AssociationErrorResponse{Status: "error", Message: fmt.Sprintf("unable to execute request, err: %v", err)}
	}

	if r.StatusCode != http.StatusCreated {
		var errorResponse AssociationErrorResponse
		err := json.Unmarshal(r.Body, &errorResponse)
		msg := "ERROR: unable to associate HubSpot objects. "
		if err != nil {
			log.Printf("%sUnable to unmarshall error response.", msg)
		} else {
			log.Printf("%sGot %d error(s):", msg, errorResponse.NumErrors)
			// HubSpot returns an error for each potential problem with the association (ie, both ids are wrong results in two errors)
			for i, itemError := range errorResponse.Errors {
				log.Printf("error %d: %s", i+1, itemError.Message)
			}
		}
		errorResponse.StatusCode = r.StatusCode
		return nil, errorResponse
	}

	var associationResult AssociationResults
	if err := json.Unmarshal(r.Body, &associationResult); err != nil {
		msg := fmt.Sprintf("could not unmarshal HubSpot response, err: %v", err)
		log.Printf("ERROR: %s", msg)
		return nil, AssociationErrorResponse{Status: "error", Message: msg}
	}

	log.Printf("INFO: HubSpot contact assocation created successfully. Contact ID: %s; Company ID %s",
		association.Inputs[0].From.ID,
		association.Inputs[0].To.ID)

	return &associationResult, AssociationErrorResponse{}
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
		log.Printf("ERROR: unable to create HubSpot contact")
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

	requestBody, err := json.Marshal(contactInput)
	if err != nil {
		log.Printf("ERROR: could not marshal the provided contact body, err: %v", err)
		return nil, ErrorResponse{Status: "error", Message: "invalid contact input"}
	}

	apiURL := fmt.Sprintf("%s/crm/%s/objects/contacts/%s?hapikey=%s", c.APIBaseURL, c.APIVersion, contactID, c.APIKey)
	r, err := c.request(apiURL, http.MethodPatch, requestBody)

	if err != nil {
		log.Printf("ERROR: unable to update HubSpot contact")
		return nil,
			ErrorResponse{Status: "error", Message: fmt.Sprintf("unable to execute request, err: %v", err)}
	}

	if r.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		err := json.Unmarshal(r.Body, &errorResponse)
		msg := "ERROR: unable to update HubSpot account:"
		if err != nil {
			log.Printf("%s unable to unmarshal error response.", msg)
		} else {
			log.Printf("%s got error: %v.", msg, errorResponse.Message)
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
