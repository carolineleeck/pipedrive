package pipedrive

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func Test_FindOrCreatePerson_Found(t *testing.T) {
	email := "test@videofruit.com"
	expectedID := 1
	client := NewClient("http://base", "abc123", ClientOptions{
		HTTPClient: fakeClient{
			reqs: map[string]string{
				"http://base/persons/find?api_token=abc123&search_by_email=1&term=test%40videofruit.com": fmt.Sprintf(personFindResp, expectedID, email),
			},
		},
	})
	person := Person{Email: []string{email}}

	err := client.FindOrCreatePerson(&person)
	if err != nil {
		t.Errorf("Unexpected error finding person: %+v", err)
		return
	}

	if person.ID != expectedID {
		t.Errorf("Failed to find person. Expected ID to be %d; got %d", expectedID, person.ID)
		return
	}
}

func Test_FindOrCreatePerson_NotFound(t *testing.T) {
	email := "test@videofruit.com"
	expectedID := 1
	client := NewClient("http://base", "abc123", ClientOptions{
		HTTPClient: fakeClient{
			reqs: map[string]string{
				"http://base/persons/find?api_token=abc123&search_by_email=1&term=test%40videofruit.com": personNoFindResp,
				"http://base/persons?api_token=abc123":                                                   fmt.Sprintf(personCreateResp, expectedID, email),
			},
		},
	})
	person := Person{Email: []string{email}}

	err := client.FindOrCreatePerson(&person)
	if err != nil {
		t.Errorf("Unexpected error finding person: %+v", err)
		return
	}

	if person.ID != expectedID {
		t.Errorf("Failed to find person. Expected ID to be %d; got %d", expectedID, person.ID)
		return
	}
}

func Test_FindOrCreateOrganization_Found(t *testing.T) {
	name := "Videofruit"
	expectedID := 1
	client := NewClient("http://base", "abc123", ClientOptions{
		HTTPClient: fakeClient{
			reqs: map[string]string{
				"http://base/organizations/find?api_token=abc123&term=Videofruit": fmt.Sprintf(orgFindResp, expectedID, name),
			},
		},
	})

	org := Organization{Name: name}

	err := client.FindOrCreateOrganization(&org)
	if err != nil {
		t.Errorf("Unexpected error finding organization: %+v", err)
		return
	}

	if org.ID != expectedID {
		t.Errorf("Failed to find Organization. Expected ID to be %d; got %d", expectedID, org.ID)
		return
	}
}
func Test_FindOrCreateOrganization_NotFound(t *testing.T) {
	expectedID := 2
	name := "Videofruit"
	client := NewClient("http://base", "abc123", ClientOptions{
		HTTPClient: fakeClient{
			reqs: map[string]string{
				"http://base/organizations/find?api_token=abc123&term=Videofruit": `{ "success": true, "data": null, "additional_data": { "pagination": { "start": 0, "limit": 100, "more_items_in_collection": false } } }`,
				"http://base/organizations?api_token=abc123":                      fmt.Sprintf(orgCreateResp, expectedID, name),
			},
		},
	})

	org := Organization{Name: name}

	err := client.FindOrCreateOrganization(&org)
	if err != nil {
		t.Errorf("Unexpected error finding or creating organization: %+v", err)
		return
	}

	if org.ID != expectedID {
		t.Errorf("Failed to create Organization. Expected ID to be %d; got %d", expectedID, org.ID)
		return
	}
}

func Test_authenticatedURLNoParams(t *testing.T) {
	base := "http://base"
	path := "/organizations"
	token := "abc123"
	client := NewClient(base, token, ClientOptions{})
	expected := base + path + "?api_token=" + token
	actual, err := client.authenticatedURL(path)
	if err != nil {
		t.Error(err)
	}
	if actual.String() != expected {
		t.Errorf("Authenticated URL want %s; got %s", expected, actual)
	}
}

func Test_authenticatedURLExistingParams(t *testing.T) {
	base := "http://base"
	param := "term=paper"
	path := "/organizations"
	token := "abc123"
	client := NewClient(base, token, ClientOptions{})
	expected := base + path + "?api_token=" + token + "&" + param
	actual, err := client.authenticatedURL(path + "?" + param)
	if err != nil {
		t.Error(err)
	}
	if actual.String() != expected {
		t.Errorf("Authenticated URL want %s; got %s", expected, actual)
	}
}

type fakeClient struct {
	reqs map[string]string
}

func (c fakeClient) Get(url string) (*http.Response, error) {
	if body, ok := c.reqs[url]; ok {
		return &http.Response{
			Body: ioutil.NopCloser(strings.NewReader(body)),
		}, nil
	}

	return nil, fmt.Errorf("URL not mocked out: %s", url)
}

func (c fakeClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	if body, ok := c.reqs[url]; ok {
		return &http.Response{
			Body: ioutil.NopCloser(strings.NewReader(body)),
		}, nil
	}

	return nil, fmt.Errorf("URL not mocked out: %s", url)
}

const orgFindResp = `{
	"success": true,
	"data": [
		{
			"id": %d,
			"name": "%s",
			"visible_to": "3"
		}
	],
	"additional_data": {
		"pagination": {
			"start": 0,
			"limit": 100,
			"more_items_in_collection": false
		}
	}
}`
const orgCreateResp = `{
	"success": true,
	"data": {
		"id": %d,
		"company_id": 2178381,
		"owner_id": {
			"id": 3219426,
			"name": "Chris Marshall",
			"email": "chris@videofruit.com",
			"has_pic": true,
			"pic_hash": "291d9e2412ec2ae669b415cf530e19d7",
			"active_flag": true,
			"value": 3219426
		},
		"name": "%s",
		"open_deals_count": 0,
		"related_open_deals_count": 0,
		"closed_deals_count": 0,
		"related_closed_deals_count": 0,
		"email_messages_count": 0,
		"people_count": 0,
		"activities_count": 0,
		"done_activities_count": 0,
		"undone_activities_count": 0,
		"reference_activities_count": 0,
		"files_count": 0,
		"notes_count": 0,
		"followers_count": 0,
		"won_deals_count": 0,
		"related_won_deals_count": 0,
		"lost_deals_count": 0,
		"related_lost_deals_count": 0,
		"active_flag": true,
		"category_id": null,
		"picture_id": null,
		"country_code": null,
		"first_char": "v",
		"update_time": "2017-11-14 17:19:21",
		"add_time": "2017-11-14 17:19:21",
		"visible_to": "3",
		"next_activity_date": null,
		"next_activity_time": null,
		"next_activity_id": null,
		"last_activity_id": null,
		"last_activity_date": null,
		"timeline_last_activity_time": null,
		"timeline_last_activity_time_by_owner": null,
		"address": null,
		"address_subpremise": null,
		"address_street_number": null,
		"address_route": null,
		"address_sublocality": null,
		"address_locality": null,
		"address_admin_area_level_1": null,
		"address_admin_area_level_2": null,
		"address_country": null,
		"address_postal_code": null,
		"address_formatted_address": null,
		"cc_email": "videofruitdev@pipedrivemail.com",
		"owner_name": "Chris Marshall",
		"edit_name": true
	},
	"related_objects": {
		"user": {
			"3219426": {
				"id": 3219426,
				"name": "Chris Marshall",
				"email": "chris@videofruit.com",
				"has_pic": true,
				"pic_hash": "291d9e2412ec2ae669b415cf530e19d7",
				"active_flag": true
			}
		}
	}
}`

const personFindResp = `{
	"success": true,
	"data": [
	{
		"id": %d,
		"name": "Some Name",
		"email": "%s",
		"phone": null,
		"org_id": null,
		"org_name": "",
		"visible_to": "3"
	}
	],
	"additional_data": {
		"search_method": "search_by_email",
		"pagination": {
			"start": 0,
			"limit": 100,
			"more_items_in_collection": false
		}
	}
}`

const personNoFindResp = `{
	"success": true,
	"data": null,
	"additional_data": {
		"search_method": "search_by_email",
		"pagination": {
			"start": 0,
			"limit": 100,
			"more_items_in_collection": false
		}
	}
}`

const personCreateResp = `{
	"success": true,
	"data": {
		"id": %d,
		"company_id": 2178381,
		"owner_id": {
			"id": 3219426,
			"name": "Chris Marshall",
			"email": "chris@videofruit.com",
			"has_pic": true,
			"pic_hash": "291d9e2412ec2ae669b415cf530e19d7",
			"active_flag": true,
			"value": 3219426
		},
		"org_id": {
			"name": "Videofruit",
			"people_count": 0,
			"owner_id": 3219426,
			"address": null,
			"cc_email": "videofruitdev@pipedrivemail.com",
			"value": 1
		},
		"name": "Tester McTest",
		"first_name": "Tester",
		"last_name": "McTest",
		"open_deals_count": 0,
		"related_open_deals_count": 0,
		"closed_deals_count": 0,
		"related_closed_deals_count": 0,
		"participant_open_deals_count": 0,
		"participant_closed_deals_count": 0,
		"email_messages_count": 0,
		"activities_count": 0,
		"done_activities_count": 0,
		"undone_activities_count": 0,
		"reference_activities_count": 0,
		"files_count": 0,
		"notes_count": 0,
		"followers_count": 0,
		"won_deals_count": 0,
		"related_won_deals_count": 0,
		"lost_deals_count": 0,
		"related_lost_deals_count": 0,
		"active_flag": true,
		"phone": [
			{
				"value": "",
				"primary": true
			}
		],
		"email": [
			{
				"label": "",
				"value": "%s",
				"primary": true
			}
		],
		"first_char": "t",
		"update_time": "2017-11-16 20:03:54",
		"add_time": "2017-11-16 20:03:54",
		"visible_to": "3",
		"picture_id": null,
		"next_activity_date": null,
		"next_activity_time": null,
		"next_activity_id": null,
		"last_activity_id": null,
		"last_activity_date": null,
		"timeline_last_activity_time": null,
		"timeline_last_activity_time_by_owner": null,
		"last_incoming_mail_time": null,
		"last_outgoing_mail_time": null,
		"org_name": "Videofruit",
		"cc_email": "videofruitdev@pipedrivemail.com",
		"owner_name": "Chris Marshall"
	},
	"related_objects": {
		"organization": {
			"1": {
				"id": 1,
				"name": "Videofruit",
				"people_count": 0,
				"owner_id": 3219426,
				"address": null,
				"cc_email": "videofruitdev@pipedrivemail.com"
			}
		},
		"user": {
			"3219426": {
				"id": 3219426,
				"name": "Chris Marshall",
				"email": "chris@videofruit.com",
				"has_pic": true,
				"pic_hash": "291d9e2412ec2ae669b415cf530e19d7",
				"active_flag": true
			}
		}
	}
}`
