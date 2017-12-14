package pipedrive

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Requestor in an interface matching http.Client
type Requestor interface {
	Get(string) (*http.Response, error)
	Post(string, string, io.Reader) (*http.Response, error)
}

// ClientOptions specifies options when creating a new Client
type ClientOptions struct {
	HTTPClient    Requestor
	DefaultUserID int
}

// Client represents a PipeDrive API client wrapper
type Client struct {
	APIToken      string
	BaseURL       string
	DefaultUserID int
	httpClient    Requestor
}

// Person is a PipeDrive Person representation
type Person struct {
	ID             int      `json:"id"`
	OwnerID        int      `json:"owner_id"`
	OrganizationID int      `json:"org_id"`
	Name           string   `json:"name"`
	Email          []string `json:"email"`
	Phone          []string `json:"phone"`
}

// Organization is a PipeDrive Organization representation
type Organization struct {
	ID      int                    `json:"id"`
	Name    string                 `json:"name"`
	OwnerID int                    `json:"owner_id"`
	Fields  map[string]interface{} `json:"fields"`
}

// Deal is a PipeDrive Deal representation
type Deal struct {
	ID             int                    `json:"id"`
	Title          string                 `json:"title"`
	Value          int                    `json:"value"`
	UserID         int                    `json:"user_id"`
	PersonID       int                    `json:"person_id"`
	OrganizationID int                    `json:"org_id"`
	Fields         map[string]interface{} `json:"fields"`
}

// NewClient returns a properly initialzed API client
func NewClient(baseURL, apiToken string, opts ClientOptions) *Client {
	client := &Client{
		APIToken:      apiToken,
		BaseURL:       baseURL,
		DefaultUserID: opts.DefaultUserID,
	}

	if opts.HTTPClient != nil {
		client.httpClient = opts.HTTPClient
	} else {
		client.httpClient = &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: time.Second * 5,
				}).Dial,
				TLSHandshakeTimeout: time.Second * 5,
			},
		}
	}

	return client
}

// FindOrCreateOrganization searches for an Organization by name and creates a
// new one if it doesn't exist
func (c *Client) FindOrCreateOrganization(org *Organization) error {
	authedURL, err := c.authenticatedURL("/organizations/find?term=" + org.Name)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Get(authedURL.String())
	if err != nil {
		return err
	}

	var data map[string]interface{}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(buf.Bytes(), &data); err != nil {
		return err
	}

	if data["data"] != nil {
		// This will likely crash us...
		org.ID = int(data["data"].([]interface{})[0].(map[string]interface{})["id"].(float64))
	} else {
		postStruct := map[string]interface{}{
			"name": org.Name,
		}
		for name, value := range org.Fields {
			postStruct[name] = value
		}

		if c.DefaultUserID != 0 {
			postStruct["owner_id"] = c.DefaultUserID
		}
		data, err := c.createEntitiy("/organizations", postStruct)
		if err != nil {
			return err
		}

		if data["data"] != nil {
			org.ID = int(data["data"].(map[string]interface{})["id"].(float64))
		} else {
			return fmt.Errorf("Error creating Pipedrive org: %s", buf.String())
		}
	}

	return nil
}

// FindOrCreatePerson creates a new Person from the initialized Person
func (c *Client) FindOrCreatePerson(newPerson *Person) error {
	if len(newPerson.Email) < 1 {
		return errors.New("Must have at least one email")
	}
	authedURL, err := c.authenticatedURL("/persons/find?search_by_email=1&term=" + newPerson.Email[0])
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Get(authedURL.String())
	if err != nil {
		return err
	}

	var data map[string]interface{}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(buf.Bytes(), &data); err != nil {
		return err
	}

	if data["data"] != nil {
		// This will likely crash us...
		newPerson.ID = int(data["data"].([]interface{})[0].(map[string]interface{})["id"].(float64))
	} else {
		postStruct := map[string]interface{}{
			"name":   newPerson.Name,
			"email":  newPerson.Email,
			"org_id": newPerson.OrganizationID,
		}
		if c.DefaultUserID != 0 {
			postStruct["owner_id"] = c.DefaultUserID
		}
		data, err := c.createEntitiy("/persons", postStruct)
		if err != nil {
			return err
		}

		// This will likely crash us similarly...
		if data["data"] != nil {
			newPerson.ID = int(data["data"].(map[string]interface{})["id"].(float64))
		} else {
			return fmt.Errorf("Error creating Pipedrive person: %s", buf.String())
		}
	}

	return nil
}

// CreateDeal creates a new Deal from the initialized Deal
func (c *Client) CreateDeal(newDeal *Deal) error {
	if c.DefaultUserID != 0 && newDeal.UserID == 0 {
		newDeal.UserID = c.DefaultUserID
	}
	bodyData := map[string]interface{}{
		"title":     newDeal.Title,
		"value":     newDeal.Value,
		"user_id":   newDeal.UserID,
		"person_id": newDeal.PersonID,
		"org_id":    newDeal.OrganizationID,
	}
	for name, value := range newDeal.Fields {
		bodyData[name] = value
	}

	data, err := c.createEntitiy("/deals", bodyData)
	if err != nil {
		return err
	}

	// Also crashing
	if data["data"] != nil {
		newDeal.ID = int(data["data"].(map[string]interface{})["id"].(float64))
	} else {
		return fmt.Errorf("Error creating Pipedrive deal: %+v", data)
	}

	return nil
}

func (c *Client) authenticatedURL(path string) (*url.URL, error) {
	authedURL, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return authedURL, err
	}

	query := authedURL.Query()
	query.Add("api_token", c.APIToken)
	authedURL.RawQuery = query.Encode()
	return authedURL, nil
}

func (c *Client) createEntitiy(path string, bodyData interface{}) (map[string]interface{}, error) {
	var data map[string]interface{}
	postBody, err := json.Marshal(bodyData)
	if err != nil {
		return data, err
	}
	postURL, err := c.authenticatedURL(path)
	if err != nil {
		return data, err
	}
	postResp, err := c.httpClient.Post(postURL.String(), "application/json", bytes.NewReader(postBody))
	if err != nil {
		return data, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(postResp.Body)
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(buf.Bytes(), &data)
	return data, err
}
