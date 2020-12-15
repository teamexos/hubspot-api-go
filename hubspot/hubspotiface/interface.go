package hubspotiface

import (
	"github.com/teamexos/hubspot-api-go/hubspot"
)

// HubSpotClient is an interface for the Hubspot Client
type HubSpotClient interface {
	CreateAssociation(association *hubspot.AssociationInput, from string, to string) (*hubspot.AssociationResults, hubspot.AssociationErrorResponse)
	CreateContact(contactInput *hubspot.ContactInput) (*hubspot.ContactOutput, hubspot.ErrorResponse)
	UpdateContact(contactID string, contactInput *hubspot.ContactInput) (*hubspot.ContactOutput, hubspot.ErrorResponse)
	ReadContact(email string, properties string) (*hubspot.ContactOutput, hubspot.ErrorResponse)
	DeleteContact(contactID string) hubspot.ErrorResponse
}

// make sure hubspot.Client type satisfies the HubSpotClient interface
var _ HubSpotClient = (*hubspot.Client)(nil)
