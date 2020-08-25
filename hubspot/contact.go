package hubspot

// ContactInput handles a contact body representation from HubSpot contact creation
type ContactInput struct {
	Properties map[string]string `json:"properties"`
}

// ContactOutput handles a contact representation from HubSpot
type ContactOutput struct {
	ID         string            `json:"id"`
	Properties map[string]string `json:"properties"`
	CreatedAt  string            `json:"created_at"`
	Date       string            `json:"date"`
	Archived   bool              `json:"archived"`
}

// NewContactInput creates a new Contact Body representation
func NewContactInput(properties map[string]string) *ContactInput {
	return &ContactInput{
		Properties: properties,
	}
}
