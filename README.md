# hubspot-api-go

v3 SDK(Client)

### Requirements

- Golang (1.14+)
- Node/npm (v15+/7+)
- Python3 w/ pip

### Getting started

```sh
make init
```

### Secret Detection

In order to ensure secure information isn't shared in the codebase and
repository, we use the [detect-secrets](https://github.com/Yelp/detect-secrets)
library to scan outgoing changes with a pre-commit hook.

In the event an inclusion of a secret is detected:

- If the secret was not a false positive, modify your code to remove the secret
and continue with your commit.
- Otherwise run `make secrets-update-baseline` to update the secrets baseline
definition and include the update with your commit.

#### Inline Allowlisting

If you want `detect-secrets` to ignore a particular line of code, simply append
an inline `pragma: allowlist secret` comment. For example:

`API_KEY = "this-looks-secret-but-is-not-secret"  # pragma: allowlist secret `

### Build and Test

```sh
make test
```

### Supported API authentication

  - API Key

### Supported API endpoints

  - Create Contact
  - Update Contact
  - Read Contact (by email address)
  - Delete Contact
  - Associate two objects (usually a contact and company)

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
