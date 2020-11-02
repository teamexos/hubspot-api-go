package hubspot

// AssociationContactToCompany is the value for the "AssociationType" when associating contact and companies
const AssociationContactToCompany = "contact_to_company"

type (
	// AssociationInput handles an association from one type of object to another
	AssociationInput struct {
		Inputs []Association `json:"inputs"`
	}

	// Association handles the two items to be associated
	Association struct {
		AssociationType string        `json:"type"`
		From            AssociationID `json:"from"`
		To              AssociationID `json:"to"`
	}

	// AssociationID handles the IDs to be associated
	AssociationID struct {
		ID string `json:"id"`
	}

	// AssociationResults handles the results from a successful call to the HubSpot Association API
	AssociationResults struct {
		Status      string        `json:"string"`
		StartedAt   string        `json:"startedAt"`
		CompletedAt string        `json:"completedAt"`
		Results     []Association `json:"results"`
	}
)

// NewSingleContactToCompanyAssociationInput can be used to connect a company to a contact
func NewSingleContactToCompanyAssociationInput(contactID string, companyID string) *AssociationInput {
	return &AssociationInput{
		Inputs: []Association{
			Association{
				AssociationType: AssociationContactToCompany,
				From: AssociationID{
					ID: contactID,
				},
				To: AssociationID{
					ID: companyID,
				},
			},
		},
	}
}
