package hubspot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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

// HubSpot Client structure
type Client struct {
	APIBaseUrl string
	APIKey     string
	APIVersion string
	HttpClient HTTPClient
}

// Creates a new HubSpot Client with corresponding defaults
func NewClient(apiKey string) *Client {
	c := &Client{}
	c.APIKey = apiKey
	c.APIBaseUrl = DefaultAPIBaseURL
	c.APIVersion = DefaultAPIVersion
	c.HttpClient = &http.Client{}
	return c
}

// Error struct returned by HubSpot API
type ErrorResponse struct {
	Category      string
	CorrelationId string
	Links         map[string]string
	Message       string
	Status        string
	StatusCode    int
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf(e.Message)
}

// Response returned by the request method
type Response struct {
	Body       json.RawMessage
	StatusCode int
}

// Creates a new Contact in HubSpot
func (c *Client) CreateContact(body Contact) (*Contact, ErrorResponse) {
	var contact Contact
	log.Printf("INFO: attempting to create HubSpot Contact")

	requestBody, err := json.Marshal(body)
	if err != nil {
		log.Printf("ERROR: could not marshal the provided contact body, err: %v", err)
		return &contact, ErrorResponse{Status: "error", Message: "invalid contact body"}
	}
	r, err := c.request(
		fmt.Sprintf("%s/crm/%s/objects/contacts/?hapikey=%s", c.APIBaseUrl, c.APIVersion, c.APIKey),
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
		msg := fmt.Sprintf("could not unmarshall HubSpot response, err: %v", err)
		log.Printf("ERROR: %s", msg)
		return &contact, ErrorResponse{Status: "error", Message: msg}
	}

	log.Printf("INFO: hubspot contact created successfully. Contact ID: %s", contact.ID)
	return &contact, ErrorResponse{}
}

// Executes a HTTP request and returns the response
func (c *Client) request(
	url string,
	method string,
	requestBody []byte) (*Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))

	var response Response
	if err != nil {
		return &response, err
	}
	req.Header.Add("Content-Type", "application/json")
	r, err := c.HttpClient.Do(req)
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
