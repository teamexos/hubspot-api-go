package hubspot

const AssocationContactToCompany = "contact_to_company"

type (
	//AssociationInput handles an association from one type of object to another
	AssociationInput struct {
		Inputs []AssociationObject `json:"inputs"`
	}

	//AssocationObject is the two objects to be associated
	AssociationObject struct {
		AssociationType string        `json:"type"`
		From            AssociationID `json:"from"`
		To              AssociationID `json:"to"`
	}

	//AssocationID is the objects's ID to be assocaiated
	AssociationID struct {
		ID string `json:"id"`
	}

	//AssociationResults are the results from a succesful call to the HubSpot Association API
	AssociationResults struct {
		Status      string              `json:"string"`
		StartedAt   string              `json:"started_at"`
		CompletedAt string              `json:"created_at"`
		Results     []AssociationObject `json:"results"`
	}
)

//NewSingleContactToCompanyAssocaiationInput can be used to connect a company to a contact
func NewSingleContactToCompanyAssociationInput(contactID string, companyID string) *AssociationInput {
	return &AssociationInput{
		Inputs: []AssociationObject{
			AssociationObject{
				AssociationType: AssocationContactToCompany,
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
