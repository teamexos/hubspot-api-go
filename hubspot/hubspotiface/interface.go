package hubspotiface

import (
	"github.com/teamexos/hubspot-api-go/hubspot"
)

// HubSpotClient is an interface for the Hubspot Client
type HubSpotClient interface {
	CreateContact(contactInput *hubspot.ContactInput) (*hubspot.ContactOutput, hubspot.ErrorResponse)
	UpdateContact(contactID string, contactInput *hubspot.ContactInput) (*hubspot.ContactOutput, hubspot.ErrorResponse)
}

// make sure hubspot.Client type satisfies the HubSpotClient interface
var _ HubSpotClient = (*hubspot.Client)(nil)
