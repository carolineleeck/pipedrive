# PipeDrive API client

The `pipedrive` package implements an API wrapper for the [PipeDrive
API](https://developers.pipedrive.com/docs/api/v1/). It exposes native go
types for common API actions.

_Status:_ We only have a create a new deal workflow implemented.
Specifically, Find or create an organization, find or create a person, and
create a deal. Other API endpoints are welcome, but not planned currently.

# Examples


```go
func demo() error {
  // Create a new client with creds
	pipeDriveClient := pipedrive.NewClient(
		os.Getenv("PIPEDRIVE_BASE_URL"),
		os.Getenv("PIPEDRIVE_API_TOKEN"),
		pipedrive.ClientOptions{
			DefaultUserID: os.Getenv("PIPEDRIVE_DEFAULT_USER"),
		},
	)
	var err error

  // Find or create an organization
	org := pipedrive.Organization{Name: "My fav org"}
	if companyName != "" {
		err = pipeDriveClient.FindOrCreateOrganization(&org)
		if err != nil {
			return err
		}
	}

  // Find or create a person
	person := pipedrive.Person{
		Name:  "Tester McTest",
		Email: []string{"test@example.com"},
	}
  // Add the organization ID if present
	if org.ID != 0 {
		person.OrganizationID = org.ID
	}
	err = pipeDriveClient.FindOrCreatePerson(&person)
	if err != nil {
		return err
	}

  // Create a deal
	deal := pipedrive.Deal{
		Title:    "Close this deal!",
		Value:    1000,
		PersonID: person.ID,
		Fields: map[string]interface{}{
			"<custom-field-hash-key>": "Custom field value",
		},
	}
  // Add the organization ID if present
	if org.ID != 0 {
		deal.OrganizationID = org.ID
	}
	err = pipeDriveClient.CreateDeal(&deal)
	if err != nil {
		return err
	}

  return nil
}
```
