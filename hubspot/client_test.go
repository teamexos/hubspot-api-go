package hubspot_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	hubSpot "github.com/teamexos/hubspot-api-go/hubspot"
	"syreclabs.com/go/faker"
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

	assert.Equal(t, "error", err.Status, "expected unauthorized error")
	assert.Equal(t, http.StatusUnauthorized, err.StatusCode, "expected 401 unauthorized error")
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
	assert.Equal(t, 0, err.StatusCode, "expected empty error response")
	assert.NotEqual(t, "", contact.ID, "expected contact id to have a value")
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
		assert.Equal(t, tt.wantStatusCode, err.StatusCode, "expected status codes to match")
		assert.Equal(t, tt.wantErrorCategory, err.Category, "expected proper error category")
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
	assert.Equal(t, 0, err.StatusCode, "expected empty error status code")
	assert.NotEqual(t, "", contact.ID, "expected a value for contact id")
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
		assert.Equal(t, tt.wantStatusCode, err.StatusCode, "expected matching error codes")
		assert.Equal(t, tt.wantErrorCategory, err.Category, "expected matching error categories")
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

	assert.Equal(t, 0, err.StatusCode, "expected empty error response, got error with status code")
	assert.Greater(t, len(association.Results), 0, "expected an array of associations")
	assert.Equal(t, contactID, association.Results[1].From.ID, "expected return contact id to match with sent value")
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

		assert.Equal(t, tt.wantStatusCode, err.StatusCode, "expected error codes to match")
		assert.Equal(t, tt.wantErrorCategory, err.Errors[0].Category, "expected error codes to match")
		assert.Len(t, err.Errors, tt.wantNumErrors, "expected number of errors")
	}
}

func TestAssociationURL(t *testing.T) {
	c := hubSpot.NewClient("invalid-api-key")

	errorMessage := "BuildAssociationURL(): from and to arguments require a value"
	tests := []struct {
		name             string
		to               string
		from             string
		expectedURL      string
		expectedErrorMsg string
	}{
		{
			name:             "two valid strings",
			to:               "company ",
			from:             " contact",
			expectedURL:      "https://api.hubapi.com/crm/v3/associations/contact/company/batch/create?hapikey=invalid-api-key",
			expectedErrorMsg: "",
		}, {
			name:             "two invalid strings",
			to:               "",
			from:             " ",
			expectedURL:      "",
			expectedErrorMsg: errorMessage,
		}, {
			name:             "one invalid string",
			to:               "",
			from:             "behold",
			expectedURL:      "",
			expectedErrorMsg: errorMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultURL, err := hubSpot.BuildAssociationURL(c, tt.from, tt.to)

			assert.Equal(t, tt.expectedURL, resultURL, "expected the proper URL for making hubspot associations")

			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorMsg, "expecting errors if to or from are blank")
			}
		})
	}

}
func TestReadContact(t *testing.T) {
	c := hubSpot.NewClient("fake-api-key")

	fakeEmail := faker.Internet().Email()

	tests := []struct {
		name             string
		json             string
		properties       string
		wantEmail        string
		wantID           string
		wantStatusCode   int
		expectedErrorMsg string
	}{
		{
			name: "found email address",
			json: fmt.Sprintf(`{
				"id": "3100",
				"properties": {
					"createdate": "2020-10-14T18:01:05.763Z",
					"email": "%s",
					"firstname": "Matt",
					"hs_object_id": "3100",
					"lastmodifieddate": "2020-10-14T18:03:10.772Z"
				},
				"createdAt": "2020-10-14T18:01:05.763Z",
				"updatedAt": "2020-10-14T18:03:10.772Z",
				"archived": false
			}`, fakeEmail),
			wantID:           "3100",
			wantEmail:        fakeEmail,
			properties:       "firstname,email",
			wantStatusCode:   http.StatusOK,
			expectedErrorMsg: "",
		},
		{
			name:             "email address not found",
			json:             "",
			wantEmail:        "mpurdon@teamexos.com",
			wantStatusCode:   http.StatusNotFound,
			properties:       "firstname,email",
			expectedErrorMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c.HTTPClient = NewMockHTTPClient(tt.wantStatusCode, tt.json)

			contactOutput, hserr := c.ReadContact(tt.wantEmail, tt.properties)

			if hserr.Status != "" {
				assert.Equal(t, http.StatusNotFound, hserr.StatusCode)
			} else {
				assert.Equal(t, tt.wantID, contactOutput.ID, "ensure the proper hubspot user ID")
				assert.Equal(t, tt.wantEmail, contactOutput.Properties["email"], "ensure the correct email address")
			}

		})
	}

}
