package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

/* =========================================================
 * Admin Client
 * ========================================================= */

type AdminClient struct {
	BaseURL      string
	Realm        string
	ClientID     string
	ClientSecret string
	httpClient   *http.Client
	token        string
}

/* =========================================================
 * Constructor
 * ========================================================= */

func NewAdminClient(
	baseURL string,
	realm string,
	clientID string,
	clientSecret string,
) (*AdminClient, error) {
	c := &AdminClient{
		BaseURL:      strings.TrimSuffix(baseURL, "/"),
		Realm:        realm,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}

	if err := c.authenticate(); err != nil {
		return nil, err
	}

	return c, nil
}

/* =========================================================
 * Authentication (client_credentials)
 * ========================================================= */

func (c *AdminClient) authenticate() error {
	tokenURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/token",
		c.BaseURL,
		c.Realm,
	)

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)

	res, err := c.httpClient.PostForm(tokenURL, form)
	if err != nil {
		return fmt.Errorf("admin auth failed: %w", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("admin auth failed [%d]: %s", res.StatusCode, body)
	}

	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &out); err != nil {
		return err
	}

	c.token = out.AccessToken
	return nil
}

/* =========================================================
 * Public API
 * ========================================================= */

// EnsureClientRoles ensures all roles exist for a given client
func (c *AdminClient) EnsureClientRoles(
	targetClientID string,
	roles []string,
) error {
	clientUUID, err := c.GetClientUUID(targetClientID)
	if err != nil {
		return err
	}

	existing, err := c.listClientRoles(clientUUID)
	if err != nil {
		return err
	}

	existingMap := map[string]bool{}
	for _, r := range existing {
		existingMap[r] = true
	}

	for _, role := range roles {
		if existingMap[role] {
			continue
		}

		if err := c.createClientRole(clientUUID, role); err != nil {
			return err
		}
	}

	return nil
}

/* =========================================================
 * Internal helpers
 * ========================================================= */

func (c *AdminClient) GetClientUUID(clientID string) (string, error) {
	u := fmt.Sprintf(
		"%s/admin/realms/%s/clients?clientId=%s",
		c.BaseURL,
		c.Realm,
		clientID,
	)

	req, _ := http.NewRequest(http.MethodGet, u, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get client failed [%d]: %s", res.StatusCode, body)
	}

	var clients []struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(body, &clients); err != nil {
		return "", err
	}

	if len(clients) == 0 {
		return "", fmt.Errorf("client %q not found in keycloak", clientID)
	}

	return clients[0].ID, nil
}

func (c *AdminClient) listClientRoles(clientUUID string) ([]string, error) {
	u := fmt.Sprintf(
		"%s/admin/realms/%s/clients/%s/roles",
		c.BaseURL,
		c.Realm,
		clientUUID,
	)

	req, _ := http.NewRequest(http.MethodGet, u, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list roles failed [%d]: %s", res.StatusCode, body)
	}

	var roles []struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal(body, &roles); err != nil {
		return nil, err
	}

	out := make([]string, 0, len(roles))
	for _, r := range roles {
		out = append(out, r.Name)
	}

	return out, nil
}

func (c *AdminClient) createClientRole(clientUUID string, role string) error {
	u := fmt.Sprintf(
		"%s/admin/realms/%s/clients/%s/roles",
		c.BaseURL,
		c.Realm,
		clientUUID,
	)

	payload := map[string]any{
		"name":        role,
		"description": "auto-provisioned by application",
	}

	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("create role %q failed [%d]: %s", role, res.StatusCode, body)
	}

	return nil
}

func (c *AdminClient) CreateRealmRoleIfMissing(
	ctx context.Context,
	roleName string,
	description string,
) error {
	// check if role exists
	checkURL := fmt.Sprintf(
		"%s/admin/realms/%s/roles/%s",
		c.BaseURL,
		c.Realm,
		roleName,
	)

	req, _ := http.NewRequest(http.MethodGet, checkURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return nil // already exists
	}

	if res.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("check realm role failed [%d]: %s", res.StatusCode, body)
	}

	// create role
	createURL := fmt.Sprintf(
		"%s/admin/realms/%s/roles",
		c.BaseURL,
		c.Realm,
	)

	payload := map[string]any{
		"name":        roleName,
		"description": description,
	}

	b, _ := json.Marshal(payload)

	req, _ = http.NewRequest(http.MethodPost, createURL, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	res, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("create realm role failed [%d]: %s", res.StatusCode, body)
	}

	return nil
}

func (c *AdminClient) CreateScopeIfMissing(
	ctx context.Context,
	clientUUID string,
	scope string,
) error {
	u := fmt.Sprintf(
		"%s/admin/realms/%s/clients/%s/authz/resource-server/scope",
		c.BaseURL,
		c.Realm,
		clientUUID,
	)

	// list scopes
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("list scopes failed [%d]: %s", res.StatusCode, body)
	}

	var scopes []struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal(body, &scopes); err != nil {
		return err
	}

	for _, s := range scopes {
		if s.Name == scope {
			return nil // exists
		}
	}

	// create scope
	payload := map[string]any{
		"name": scope,
	}

	b, _ := json.Marshal(payload)

	req, _ = http.NewRequest(http.MethodPost, u, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	res, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("create scope failed [%d]: %s", res.StatusCode, body)
	}

	return nil
}

func (c *AdminClient) CreateResourceIfMissing(
	ctx context.Context,
	clientUUID string,
	resource string,
) error {
	u := fmt.Sprintf(
		"%s/admin/realms/%s/clients/%s/authz/resource-server/resource",
		c.BaseURL,
		c.Realm,
		clientUUID,
	)

	// list resources
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("list resources failed [%d]: %s", res.StatusCode, body)
	}

	var resources []struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal(body, &resources); err != nil {
		return err
	}

	for _, r := range resources {
		if r.Name == resource {
			return nil // exists
		}
	}

	// create resource
	payload := map[string]any{
		"name": resource,
		"type": resource,
	}

	b, _ := json.Marshal(payload)

	req, _ = http.NewRequest(http.MethodPost, u, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	res, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("create resource failed [%d]: %s", res.StatusCode, body)
	}

	return nil
}

func (c *AdminClient) CreatePermissionIfMissing(
	ctx context.Context,
	clientUUID string,
	permissionName string,
	resourceName string,
	scopes []string,
) error {

	// --------------------------------------------------
	// 1. List existing permissions
	// --------------------------------------------------
	listURL := fmt.Sprintf(
		"%s/admin/realms/%s/clients/%s/authz/resource-server/permission",
		c.BaseURL,
		c.Realm,
		clientUUID,
	)

	req, _ := http.NewRequest(http.MethodGet, listURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("list permissions failed [%d]: %s", res.StatusCode, body)
	}

	var perms []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal(body, &perms); err != nil {
		return err
	}

	for _, p := range perms {
		if p.Name == permissionName {
			return nil // already exists
		}
	}

	// --------------------------------------------------
	// 2. Resolve resource ID
	// --------------------------------------------------
	resourceURL := fmt.Sprintf(
		"%s/admin/realms/%s/clients/%s/authz/resource-server/resource",
		c.BaseURL,
		c.Realm,
		clientUUID,
	)

	req, _ = http.NewRequest(http.MethodGet, resourceURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ = io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("list resources failed [%d]: %s", res.StatusCode, body)
	}

	var resources []struct {
		ID   string `json:"_id"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal(body, &resources); err != nil {
		return err
	}

	var resourceID string
	for _, r := range resources {
		if r.Name == resourceName {
			resourceID = r.ID
			break
		}
	}

	if resourceID == "" {
		return fmt.Errorf("resource %q not found (create resource first)", resourceName)
	}

	// --------------------------------------------------
	// 3. Create permission
	// --------------------------------------------------
	createURL := fmt.Sprintf(
		"%s/admin/realms/%s/clients/%s/authz/resource-server/permission/resource",
		c.BaseURL,
		c.Realm,
		clientUUID,
	)

	payload := map[string]any{
		"name":             permissionName,
		"type":             "resource",
		"logic":            "POSITIVE",
		"decisionStrategy": "UNANIMOUS",
		"resources":        []string{resourceID},
		"scopes":           scopes,
	}

	b, _ := json.Marshal(payload)

	req, _ = http.NewRequest(http.MethodPost, createURL, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	res, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("create permission failed [%d]: %s", res.StatusCode, body)
	}

	return nil
}

func (c *AdminClient) AttachPolicyToPermission(
	ctx context.Context,
	clientUUID string,
	permissionName string,
	policyName string,
) error {
	// list permissions
	listURL := fmt.Sprintf(
		"%s/admin/realms/%s/clients/%s/authz/resource-server/permission",
		c.BaseURL,
		c.Realm,
		clientUUID,
	)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("list permissions failed [%d]: %s", res.StatusCode, body)
	}

	var perms []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal(body, &perms); err != nil {
		return err
	}

	var permID string
	for _, p := range perms {
		if p.Name == permissionName {
			permID = p.ID
			break
		}
	}

	if permID == "" {
		return fmt.Errorf("permission %q not found", permissionName)
	}

	// get permission
	getURL := fmt.Sprintf(
		"%s/admin/realms/%s/clients/%s/authz/resource-server/permission/%s",
		c.BaseURL,
		c.Realm,
		clientUUID,
		permID,
	)

	req, _ = http.NewRequestWithContext(ctx, http.MethodGet, getURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var perm map[string]any
	if err := json.NewDecoder(res.Body).Decode(&perm); err != nil {
		return err
	}

	// normalize policies list
	rawPolicies, ok := perm["policies"].([]any)
	if !ok {
		rawPolicies = []any{}
	}

	for _, p := range rawPolicies {
		if s, ok := p.(string); ok && s == policyName {
			return nil // already attached
		}
	}

	rawPolicies = append(rawPolicies, policyName)
	perm["policies"] = rawPolicies

	b, _ := json.Marshal(perm)

	req, _ = http.NewRequestWithContext(ctx, http.MethodPut, getURL, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	res, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// ✅ accept all 2xx responses
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("attach policy failed [%d]: %s", res.StatusCode, body)
	}

	return nil
}

func (c *AdminClient) CreateRolePolicyIfMissing(
	ctx context.Context,
	clientUUID string,
	policyName string,
	roleName string,
) error {
	baseURL := fmt.Sprintf(
		"%s/admin/realms/%s/clients/%s/authz/resource-server/policy/role",
		c.BaseURL,
		c.Realm,
		clientUUID,
	)

	// --------------------------------------------------
	// 1. List existing role policies
	// --------------------------------------------------
	req, _ := http.NewRequest(http.MethodGet, baseURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("list role policies failed [%d]: %s", res.StatusCode, body)
	}

	var policies []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal(body, &policies); err != nil {
		return err
	}

	for _, p := range policies {
		if p.Name == policyName {
			return nil // already exists
		}
	}

	// --------------------------------------------------
	// 2. Create role policy
	// --------------------------------------------------
	payload := map[string]any{
		"name":             policyName,
		"description":      "auto-provisioned by application",
		"type":             "role",
		"logic":            "POSITIVE",
		"decisionStrategy": "UNANIMOUS",
		"roles": []map[string]any{
			{
				"id":       roleName,
				"required": false,
			},
		},
	}

	b, _ := json.Marshal(payload)

	req, _ = http.NewRequest(http.MethodPost, baseURL, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	res, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("create role policy failed [%d]: %s", res.StatusCode, body)
	}

	return nil
}

func (c *AdminClient) AssignRealmRole(
	ctx context.Context,
	userID string,
	roleName string,
) error {

	// --------------------------------------------------
	// 1. Get realm role representation
	// --------------------------------------------------
	roleURL := fmt.Sprintf(
		"%s/admin/realms/%s/roles/%s",
		c.BaseURL,
		c.Realm,
		roleName,
	)

	req, _ := http.NewRequest(http.MethodGet, roleURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("get realm role failed [%d]: %s", res.StatusCode, body)
	}

	var role map[string]any
	if err := json.Unmarshal(body, &role); err != nil {
		return err
	}

	// --------------------------------------------------
	// 2. Assign role to user
	// --------------------------------------------------
	assignURL := fmt.Sprintf(
		"%s/admin/realms/%s/users/%s/role-mappings/realm",
		c.BaseURL,
		c.Realm,
		userID,
	)

	payload := []map[string]any{
		{
			"id":   role["id"],
			"name": role["name"],
		},
	}

	b, _ := json.Marshal(payload)

	req, _ = http.NewRequest(http.MethodPost, assignURL, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	res, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("assign realm role failed [%d]: %s", res.StatusCode, body)
	}

	return nil
}

func (c *AdminClient) AssignClientRole(
	userID string,
	targetClientID string,
	roleName string,
) error {

	// --------------------------------------------------
	// 1. Resolve client UUID
	// --------------------------------------------------
	clientUUID, err := c.GetClientUUID(targetClientID)
	if err != nil {
		return err
	}

	// --------------------------------------------------
	// 2. Fetch client role representation
	// --------------------------------------------------
	roleURL := fmt.Sprintf(
		"%s/admin/realms/%s/clients/%s/roles/%s",
		c.BaseURL,
		c.Realm,
		clientUUID,
		roleName,
	)

	req, _ := http.NewRequest(http.MethodGet, roleURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("get client role failed [%d]: %s", res.StatusCode, body)
	}

	var role map[string]any
	if err := json.Unmarshal(body, &role); err != nil {
		return err
	}

	// --------------------------------------------------
	// 3. Assign role to user
	// --------------------------------------------------
	assignURL := fmt.Sprintf(
		"%s/admin/realms/%s/users/%s/role-mappings/clients/%s",
		c.BaseURL,
		c.Realm,
		userID,
		clientUUID,
	)

	payload := []map[string]any{
		{
			"id":   role["id"],
			"name": role["name"],
		},
	}

	b, _ := json.Marshal(payload)

	req, _ = http.NewRequest(http.MethodPost, assignURL, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	res, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("assign client role failed [%d]: %s", res.StatusCode, body)
	}

	return nil
}

func (c *AdminClient) EnsureUser(
	ctx context.Context,
	username string,
	email string,
	firstName string,
	lastName string,
	enabled bool,
) (string, error) {

	// --------------------------------------------------
	// 1. Check if user already exists
	// --------------------------------------------------
	searchURL := fmt.Sprintf(
		"%s/admin/realms/%s/users?username=%s",
		c.BaseURL,
		c.Realm,
		url.QueryEscape(username),
	)

	req, _ := http.NewRequest(http.MethodGet, searchURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search user failed [%d]: %s", res.StatusCode, body)
	}

	var users []struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}

	if err := json.Unmarshal(body, &users); err != nil {
		return "", err
	}

	if len(users) > 0 {
		return users[0].ID, nil // user exists
	}

	// --------------------------------------------------
	// 2. Create user
	// --------------------------------------------------
	createURL := fmt.Sprintf(
		"%s/admin/realms/%s/users",
		c.BaseURL,
		c.Realm,
	)

	payload := map[string]any{
		"username":  username,
		"email":     email,
		"firstName": firstName,
		"lastName":  lastName,
		"enabled":   enabled,
	}

	b, _ := json.Marshal(payload)

	req, _ = http.NewRequest(http.MethodPost, createURL, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	res, err = c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("create user failed [%d]: %s", res.StatusCode, body)
	}

	// --------------------------------------------------
	// 3. Fetch newly created user ID
	// --------------------------------------------------
	req, _ = http.NewRequest(http.MethodGet, searchURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err = c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var created []struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(res.Body).Decode(&created); err != nil {
		return "", err
	}

	if len(created) == 0 {
		return "", fmt.Errorf("user created but not found")
	}

	return created[0].ID, nil
}

func (c *AdminClient) EnsureClientRoleAssigned(
	userID string,
	targetClientID string,
	roleName string,
) error {

	clientUUID, err := c.GetClientUUID(targetClientID)
	if err != nil {
		return err
	}

	// --------------------------------------------------
	// Check existing assignments
	// --------------------------------------------------
	listURL := fmt.Sprintf(
		"%s/admin/realms/%s/users/%s/role-mappings/clients/%s",
		c.BaseURL,
		c.Realm,
		userID,
		clientUUID,
	)

	req, _ := http.NewRequest(http.MethodGet, listURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("list client role mappings failed [%d]: %s", res.StatusCode, body)
	}

	var remember []struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal(body, &remember); err != nil {
		return err
	}

	for _, r := range remember {
		if r.Name == roleName {
			return nil // already assigned
		}
	}

	// --------------------------------------------------
	// Assign if missing
	// --------------------------------------------------
	return c.AssignClientRole(userID, targetClientID, roleName)
}
