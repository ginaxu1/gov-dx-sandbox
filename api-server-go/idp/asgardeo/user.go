package asgardeo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gov-dx-sandbox/api-server-go/idp"
)

type GetUserResponseBody struct {
	ID       string   `json:"id"`
	UserName string   `json:"userName"`
	Email    []string `json:"emails"`
	Name     struct {
		FamilyName string `json:"familyName"`
		GivenName  string `json:"givenName"`
	} `json:"name"`
}
type CreateUserRequestBodyEmail struct {
	Value   string `json:"value"`
	Primary bool   `json:"primary"`
}
type CreateUserRequestBody struct {
	UserName string                       `json:"userName"`
	Email    string                       `json:"email"`
	Emails   []CreateUserRequestBodyEmail `json:"emails"`
	Name     struct {
		FamilyName string `json:"familyName"`
		GivenName  string `json:"givenName"`
	} `json:"name"`
	Schema interface{} `json:"urn:scim:wso2:schema"`
}

type RoleType struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type CreateUserResponseBody struct {
	ID   string `json:"id"`
	Name struct {
		FamilyName string `json:"familyName"`
		GivenName  string `json:"givenName"`
	} `json:"name"`
	Emails   []string   `json:"emails"`
	Roles    []RoleType `json:"roles"`
	UserName string     `json:"username"`
}

func (a *Client) GetUser(ctx context.Context, userId string) (*idp.UserInfo, error) {
	url := fmt.Sprintf("%s/scim2/Users/%s", a.BaseURL, userId)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := a.Client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get user, status code: %d", res.StatusCode)
	}

	var response GetUserResponseBody

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	userInfo := &idp.UserInfo{
		Id:        response.ID,
		FirstName: response.Name.GivenName,
		LastName:  response.Name.FamilyName,
	}

	if len(response.Email) > 0 {
		userInfo.Email = response.Email[0]
	}

	return userInfo, nil
}

func (a *Client) CreateUser(ctx context.Context, userInfo *idp.User) (*idp.UserInfo, error) {
	url := fmt.Sprintf("%s/scim2/Users", a.BaseURL)

	body := CreateUserRequestBody{
		UserName: fmt.Sprintf("DEFAULT/%s", userInfo.Email),
		Email:    userInfo.Email,
		Emails: []CreateUserRequestBodyEmail{
			{
				Value:   userInfo.Email,
				Primary: true,
			},
		},
		Schema: map[string]interface{}{
			"askPassword": true,
		},
	}

	body.Name.GivenName = userInfo.FirstName
	body.Name.FamilyName = userInfo.LastName

	payload, err := json.Marshal(body)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/scim+json")

	res, err := a.Client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if res.StatusCode != 201 {
		return nil, fmt.Errorf("failed to create user, status code: %d", res.StatusCode)
	}

	var response CreateUserResponseBody

	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	createdInfo := &idp.UserInfo{
		Id:        response.ID,
		FirstName: response.Name.GivenName,
		LastName:  response.Name.FamilyName,
	}

	if len(response.Emails) > 0 {
		createdInfo.Email = response.Emails[0]
	}

	return createdInfo, nil
}

func (a *Client) DeleteUser(ctx context.Context, userId string) error {
	url := fmt.Sprintf("%s/scim2/Users/%s", a.BaseURL, userId)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	res, err := a.Client.Do(req)

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	if res.StatusCode != 204 {
		return fmt.Errorf("failed to delete user, status code: %d", res.StatusCode)
	}

	return nil
}
