package hubspot

// Contact handles a contact representation from HubSpot
type Contact struct {
	ID         string            `json:"id"`
	Properties map[string]string `json:"properties"`
}

// NewContact creates a new Contact Body representation
// TODO: I'll change this so it's more flexible and allows different properties
func NewContact(
	firstName string,
	lastName string,
	email string,
	workEmail string,
	company string) *Contact {
	return &Contact{
		Properties: map[string]string{
			"firstname":  firstName,
			"lastname":   lastName,
			"email":      email,
			"work_email": workEmail,
			"company":    company,
		},
	}
}
