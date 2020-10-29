package hubspot_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	hubSpot "github.com/teamexos/hubspot-api-go/hubspot"
)

type MockHTTPClient struct {
	wantResponse     string
	wantResponseCode int
	DoFunc           func(req *http.Request) (*http.Response, error)
}

func NewMockHTTPClient(wantResponseCode int, wantResponse string) *MockHTTPClient {
	return &MockHTTPClient{
		wantResponse:     wantResponse,
		wantResponseCode: wantResponseCode,
	}
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte(m.wantResponse)))
	return &http.Response{
		StatusCode: m.wantResponseCode,
		Body:       r,
	}, nil
}

func TestUnauthorized(t *testing.T) {
	c := hubSpot.NewClient("invalid-api-key!")
	c.HTTPClient = NewMockHTTPClient(
		http.StatusUnauthorized,
		`{
		"status": "error",
			"message": "The API key provided is invalid.",
			"correlationId": "2af0c5ea-1cb7-438e-8e60-37e8ea6879d5",
			"category": "INVALID_AUTHENTICATION",
			"links": {
			"api key": "https://app.hubspot.com/l/api-key/"
		}
	}`)
	_, err := c.CreateContact(hubSpot.NewContactInput(map[string]string{}))
	if err.Status != "error" {
		t.Errorf("expected unauthorized error, got: %s", err)
	}
	if err.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 unauthorized status error, got: %d", err.StatusCode)
	}
}

func TestCreateContact(t *testing.T) {
	c := hubSpot.NewClient("this-Is-A-Secret-!")

	properties := map[string]string{
		"firstname":  "Peter",
		"lastname":   "Parker",
		"email":      "pp@gmail.com",
		"work_email": "pp@marvel.com",
		"company":    "Marvel",
	}

	c.HTTPClient = NewMockHTTPClient(
		http.StatusCreated,
		`{
			"id": "551",
			"properties": {
				"company": "Marvel",
				"createdate": "2020-08-20T15:47:54.554Z",
				"email": "pp@gmail.com",
				"firstname": "Peter",
				"hs_is_unworked": "true",
				"lastmodifieddate": "2020-08-20T15:47:54.870Z",
				"lastname": "Parker"
			},
			"createdAt": "2020-08-20T15:47:54.554Z",
			"updatedAt": "2020-08-20T15:47:54.870Z",
			"archived": false
		}`)

	contact, err := c.CreateContact(hubSpot.NewContactInput(properties))
	if err.StatusCode != 0 {
		t.Errorf("expected empty error response, got error with status code: %d", err.StatusCode)
	}

	if contact.ID == "" {
		t.Errorf("expected a contact ID different than empty")
	}
}

func TestCreateContactErrors(t *testing.T) {
	//c := hubSpot.NewClient("this-Is-A-Secret-!")
	c := hubSpot.NewClient("11a17991-a99a-4cf3-93f1-c7ed2345f941")

	properties := map[string]string{
		"firstname":  "Peter",
		"lastname":   "Parker",
		"email":      "pp@gmail.com",
		"work_email": "pp@marvel.com",
		"company":    "Marvel",
	}
	wantContact := hubSpot.NewContactInput(properties)

	tests := []struct {
		name              string
		json              string
		wantStatusCode    int
		wantErrorCategory string
	}{
		{
			name: "contactAlreadyExists",
			json: `{
				"status": "error",
				"message": "Contact already exists",
				"correlationId": "64c72d80-c369-409f-b2ec-c233d4928080",
				"category": "CONFLICT"
			}`,
			wantStatusCode:    http.StatusConflict,
			wantErrorCategory: "CONFLICT",
		},
		{
			name: "badRequest",
			json: `{
				"status": "error",
				"message": "Property values were not valid",
				"correlationId": "cfa4f261-2877-4f61-8a75-e411c5163134",
				"category": "VALIDATION_ERROR"
			}`,
			wantStatusCode:    http.StatusBadRequest,
			wantErrorCategory: "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		c.HTTPClient = NewMockHTTPClient(
			tt.wantStatusCode,
			tt.json,
		)
		_, err := c.CreateContact(wantContact)
		if err.StatusCode != tt.wantStatusCode {
			t.Errorf("expected error code: %d, got : %d", tt.wantStatusCode, err.StatusCode)
		}
		if err.Category != tt.wantErrorCategory {
			t.Errorf("expected error category: %s, got : %s", tt.wantErrorCategory, err.Category)
		}
	}
}

func TestUpdateContact(t *testing.T) {
	c := hubSpot.NewClient("this-Is-A-Secret-!")

	properties := map[string]string{
		"exos_perform_account_verified": "true",
		"company":                       "Marvel",
	}

	c.HTTPClient = NewMockHTTPClient(
		http.StatusOK,
		`{
			"id": "551",
			"properties": {
				"company": "Marvel",
				"createdate": "2020-08-20T15:47:54.554Z",
				"email": "pp@gmail.com",
				"firstname": "Peter",
				"hs_is_unworked": "true",
				"lastmodifieddate": "2020-08-20T15:47:54.870Z",
				"exos_perform_account_verified": "true",
				"lastname": "Parker"
			},
			"createdAt": "2020-08-20T15:47:54.554Z",
			"updatedAt": "2020-08-20T15:47:54.870Z",
			"archived": false
		}`)

	contact, err := c.UpdateContact("ContactID", hubSpot.NewContactInput(properties))
	if err.StatusCode != 0 {
		t.Errorf("expected empty error response, got error with status code: %d", err.StatusCode)
	}

	if contact.ID == "" {
		t.Errorf("expected a contact ID different than empty")
	}
}

func TestUpdateContactErrors(t *testing.T) {
	c := hubSpot.NewClient("this-Is-A-Secret-!")

	properties := map[string]string{
		"exos_perform_account_verified": "true",
		"company":                       "Marvel",
	}
	wantContact := hubSpot.NewContactInput(properties)

	tests := []struct {
		name              string
		json              string
		wantStatusCode    int
		wantErrorCategory string
	}{
		{
			name: "badRequest",
			json: `{
				"status": "error",
				"message": "Property values were not valid",
				"correlationId": "cfa4f261-2877-4f61-8a75-e411c5163134",
				"category": "VALIDATION_ERROR"
			}`,
			wantStatusCode:    http.StatusBadRequest,
			wantErrorCategory: "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		c.HTTPClient = NewMockHTTPClient(
			tt.wantStatusCode,
			tt.json,
		)
		contactId := "someContactId"
		_, err := c.UpdateContact(contactId, wantContact)
		if err.StatusCode != tt.wantStatusCode {
			t.Errorf("expected error code: %d, got : %d", tt.wantStatusCode, err.StatusCode)
		}
		if err.Category != tt.wantErrorCategory {
			t.Errorf("expected error category: %s, got : %s", tt.wantErrorCategory, err.Category)
		}
	}
}

func TestCreateAssociation(t *testing.T) {
	c := hubSpot.NewClient("this-Is-A-Secret-!")

	c.HTTPClient = NewMockHTTPClient(
		http.StatusCreated,
		`{
			"status": "COMPLETE",
			"results": [
				{
					"from": {
						"id": "4705054985"
					},
					"to": {
						"id": "3051"
					},
					"type": "company_to_contact"
				},
				{
					"from": {
						"id": "3051"
					},
					"to": {
						"id": "4705054985"
					},
					"type": "contact_to_company"
				}
			],
			"startedAt": "2020-10-29T12:32:11.425Z",
			"completedAt": "2020-10-29T12:32:11.451Z"
		}`)

	contactID := "3051"
	companyID := "4705054985"
	wantAssociation := hubSpot.NewSingleContactToCompanyAssociationInput(contactID, companyID)
	association, err := c.CreateAssociation(wantAssociation, "contact", "company")
	if err.StatusCode != 0 {
		t.Errorf("expected empty error response, got error with status code: %d", err.StatusCode)
	}

	if len(association.Results) == 0 {
		t.Errorf("expected an array of assocations")
	}

	if association.Results[1].From.ID != contactID {
		t.Errorf("expected return contact id to match with sent value")
	}
}

func TestCreateAssociationErrors(t *testing.T) {
	c := hubSpot.NewClient("this-Is-A-Secret-!")

	contactID := "3051"
	companyID := "4705054985"
	wantAssociation := hubSpot.NewSingleContactToCompanyAssociationInput(contactID, companyID)

	tests := []struct {
		name              string
		json              string
		wantStatusCode    int
		wantErrorCategory string
		wantNumErrors     int
	}{
		{
			name: "invalidContact",
			json: `{
				"status": "COMPLETE",
				"results": [],
				"numErrors": 1,
				"errors": [
					{
						"status": "error",
						"category": "OBJECT_NOT_FOUND",
						"subCategory": "crm.associations.FROM_OBJECT_NOT_FOUND",
						"message": "No contact with ID 9993051 exists",
						"context": {
							"objectType": [
								"contact"
							],
							"id": [
								"9993051"
							]
						}
					}
				],
				"startedAt": "2020-10-29T12:43:31.395Z",
				"completedAt": "2020-10-29T12:43:31.404Z"
			}`,
			wantStatusCode:    http.StatusMultiStatus,
			wantErrorCategory: "OBJECT_NOT_FOUND",
			wantNumErrors:     1,
		},
		{
			name: "invalidContactAndCompany",
			json: `{
				"status": "COMPLETE",
				"results": [],
				"numErrors": 2,
				"errors": [
					{
						"status": "error",
						"category": "OBJECT_NOT_FOUND",
						"subCategory": "crm.associations.FROM_OBJECT_NOT_FOUND",
						"message": "No contact with ID 9993051 exists",
						"context": {
							"objectType": [
								"contact"
							],
							"id": [
								"9993051"
							]
						}
					},
					{
						"status": "error",
						"category": "OBJECT_NOT_FOUND",
						"subCategory": "crm.associations.TO_OBJECT_NOT_FOUND",
						"message": "No company with ID 994705054985 exists",
						"context": {
							"objectType": [
								"company"
							],
							"id": [
								"994705054985"
							]
						}
					}
				],
				"startedAt": "2020-10-29T12:58:30.125Z",
				"completedAt": "2020-10-29T12:58:30.133Z"
			}`,
			wantStatusCode:    http.StatusMultiStatus,
			wantErrorCategory: "OBJECT_NOT_FOUND",
			wantNumErrors:     2,
		},
	}

	for _, tt := range tests {
		c.HTTPClient = NewMockHTTPClient(
			tt.wantStatusCode,
			tt.json,
		)
		_, err := c.CreateAssociation(wantAssociation, "contact", "company")
		if err.StatusCode != tt.wantStatusCode {
			t.Errorf("expected error code: %d, got : %d", tt.wantStatusCode, err.StatusCode)
		}
		if err.Errors[0].Category != tt.wantErrorCategory {
			t.Errorf("expected error category: %s, got : %s", tt.wantErrorCategory, err.Errors[0].Category)
		}

		if len(err.Errors) != tt.wantNumErrors {
			t.Errorf("expected %d errors, but received %d", tt.wantNumErrors, len(err.Errors))
		}
	}
}
