# hubspot-api-go

v3 SDK(Client) 

### Requirements

- Golang (1.14+)
  - `dep` (0.5.4)
  
### Build and Test

```sh
make build
make test
```

### Supported API authentication

  - API Key
  
### Supported API endpoints

  - Create Contact
  
## Usage

```go
package main

import (
	"fmt"
hubSpot "github.com/teamexos/hubspot-api-go/hubspot"
)

func main(){
    // with api_key
    client := hubSpot.NewClient("your_api_key")
    newContact := hubSpot.NewContact(
        "Peter",
        "Parker",
        "pp@gmail.com",
        "pp@marvel.com",
        "Marvel")
    
    contact, err := client.CreateContact(newContact)
    if err.Status != "" {
        fmt.Printf("Contact ID: %s", contact.ID)
    }
}
```
