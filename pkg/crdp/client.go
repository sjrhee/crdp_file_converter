package crdp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a CRDP API client that communicates with CRDP server
type Client struct {
	baseURL  string
	policy   string
	timeout  time.Duration
	client   *http.Client
}

// APIResponse represents a response from the CRDP API
type APIResponse struct {
	StatusCode int
	Body       interface{}
	RequestURL string
	Error      error
}

// IsSuccess returns true if the API response indicates success
func (r *APIResponse) IsSuccess() bool {
	return r.Error == nil && r.StatusCode == 200
}

// ProtectBulkRequest is a bulk data protection request
type ProtectBulkRequest struct {
	ProtectionPolicyName string   `json:"protection_policy_name"`
	DataArray            []string `json:"data_array"`
	Username             string   `json:"username,omitempty"`
}

// RevealBulkRequest is a bulk data reveal request
type RevealBulkRequest struct {
	ProtectionPolicyName string              `json:"protection_policy_name"`
	ProtectedDataArray   []map[string]string `json:"protected_data_array"`
	Username             string              `json:"username,omitempty"`
}

// NewClient creates a new CRDP API client with the given configuration
func NewClient(host string, port int, policy string, timeout int) *Client {
	baseURL := fmt.Sprintf("http://%s:%d", host, port)

	httpClient := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	return &Client{
		baseURL:  baseURL,
		policy:   policy,
		timeout:  time.Duration(timeout) * time.Second,
		client:   httpClient,
	}
}

// Close closes the client and frees resources
func (c *Client) Close() error {
	c.client.CloseIdleConnections()
	return nil
}

// postJSON sends a JSON POST request and returns the response
func (c *Client) postJSON(endpoint string, payload interface{}) *APIResponse {
	fullURL := c.baseURL + endpoint
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return &APIResponse{
			StatusCode: 0,
			Body:       nil,
			RequestURL: fullURL,
			Error:      err,
		}
	}

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return &APIResponse{
			StatusCode: 0,
			Body:       nil,
			RequestURL: fullURL,
			Error:      err,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.client.Do(req)
	if err != nil {
		return &APIResponse{
			StatusCode: 0,
			Body:       nil,
			RequestURL: fullURL,
			Error:      err,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIResponse{
			StatusCode: resp.StatusCode,
			Body:       nil,
			RequestURL: fullURL,
			Error:      err,
		}
	}

	var bodyInterface interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &bodyInterface); err != nil {
			// If JSON parsing fails, store as string
			bodyInterface = string(body)
		}
	} else {
		bodyInterface = ""
	}

	return &APIResponse{
		StatusCode: resp.StatusCode,
		Body:       bodyInterface,
		RequestURL: fullURL,
		Error:      nil,
	}
}

// ProtectBulk protects multiple data
func (c *Client) ProtectBulk(dataList []string) *APIResponse {
	payload := ProtectBulkRequest{
		ProtectionPolicyName: c.policy,
		DataArray:            dataList,
	}
	return c.postJSON("/v1/protectbulk", payload)
}

// RevealBulk reveals multiple protected data
func (c *Client) RevealBulk(protectedDataList []string) *APIResponse {
	protectedDataArray := make([]map[string]string, len(protectedDataList))
	for i, pd := range protectedDataList {
		protectedDataArray[i] = map[string]string{"protected_data": pd}
	}
	
	payload := RevealBulkRequest{
		ProtectionPolicyName: c.policy,
		ProtectedDataArray:   protectedDataArray,
	}
	return c.postJSON("/v1/revealbulk", payload)
}

// ExtractProtectedFromProtectResponse extracts protected data from protect response
func (c *Client) ExtractProtectedFromProtectResponse(resp *APIResponse) string {
	if resp == nil || resp.Body == nil {
		return ""
	}

	bodyMap, ok := resp.Body.(map[string]interface{})
	if !ok {
		return ""
	}

	// Try different possible keys
	keys := []string{"protected_data", "protected", "token"}
	for _, key := range keys {
		if val, exists := bodyMap[key]; exists {
			if strVal, ok := val.(string); ok {
				return strVal
			}
		}
	}

	return ""
}

// ExtractRestoredFromRevealResponse extracts restored data from reveal response
func (c *Client) ExtractRestoredFromRevealResponse(resp *APIResponse) string {
	if resp == nil || resp.Body == nil {
		return ""
	}

	bodyMap, ok := resp.Body.(map[string]interface{})
	if !ok {
		return ""
	}

	// Try different possible keys
	keys := []string{"data", "original", "plain", "revealed", "unprotected_data", "unprotected", "decrypted"}
	for _, key := range keys {
		if val, exists := bodyMap[key]; exists {
			if strVal, ok := val.(string); ok {
				return strVal
			}
		}
	}

	return ""
}

// ExtractProtectedListFromProtectResponse extracts list of protected data from bulk protect response
func (c *Client) ExtractProtectedListFromProtectResponse(resp *APIResponse) []string {
	if resp == nil || resp.Body == nil {
		return []string{}
	}

	// Try list directly
	if list, ok := resp.Body.([]interface{}); ok {
		return c.convertToStringList(list)
	}

	// Try dictionary response
	bodyMap, ok := resp.Body.(map[string]interface{})
	if !ok {
		return []string{}
	}

	return c.extractProtectedListFromDict(bodyMap)
}

// ExtractRestoredListFromRevealResponse extracts list of restored data from bulk reveal response
func (c *Client) ExtractRestoredListFromRevealResponse(resp *APIResponse) []string {
	if resp == nil || resp.Body == nil {
		return []string{}
	}

	// Try list directly
	if list, ok := resp.Body.([]interface{}); ok {
		return c.convertToStringList(list)
	}

	// Try dictionary response
	bodyMap, ok := resp.Body.(map[string]interface{})
	if !ok {
		return []string{}
	}

	return c.extractRestoredListFromDict(bodyMap)
}

func (c *Client) extractProtectedListFromDict(bodyMap map[string]interface{}) []string {
	// Check for top-level list
	if val, exists := bodyMap["protected_data"]; exists {
		if list, ok := val.([]interface{}); ok {
			return c.convertToStringList(list)
		}
	}

	// Check for Thales style: protected_data_array
	if val, exists := bodyMap["protected_data_array"]; exists {
		if list, ok := val.([]interface{}); ok {
			return c.extractFromArray(list, "protected_data")
		}
	}

	// Check for results list
	if val, exists := bodyMap["results"]; exists {
		if list, ok := val.([]interface{}); ok {
			return c.extractFromArray(list, "protected_data")
		}
	}

	return []string{}
}

func (c *Client) extractRestoredListFromDict(bodyMap map[string]interface{}) []string {
	// Check for direct list keys
	keys := []string{"data", "restored", "items"}
	for _, key := range keys {
		if val, exists := bodyMap[key]; exists {
			if list, ok := val.([]interface{}); ok {
				return c.convertToStringList(list)
			}
		}
	}

	// Check for results list
	if val, exists := bodyMap["results"]; exists {
		if list, ok := val.([]interface{}); ok {
			return c.extractRestoredFromResults(list)
		}
	}

	// Check for Thales style: data_array
	if val, exists := bodyMap["data_array"]; exists {
		if list, ok := val.([]interface{}); ok {
			return c.extractFromArray(list, "data")
		}
	}

	// Fallback: extract string values
	result := []string{}
	for _, v := range bodyMap {
		if strVal, ok := v.(string); ok {
			result = append(result, strVal)
		}
	}
	return result
}

func (c *Client) extractFromArray(items []interface{}, key string) []string {
	result := []string{}
	for _, item := range items {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if val, exists := itemMap[key]; exists {
				if strVal, ok := val.(string); ok {
					result = append(result, strVal)
				}
			}
		}
	}
	return result
}

func (c *Client) extractRestoredFromResults(results []interface{}) []string {
	result := []string{}
	keys := []string{"data", "restored", "value"}
	
	for _, item := range results {
		if itemMap, ok := item.(map[string]interface{}); ok {
			for _, key := range keys {
				if val, exists := itemMap[key]; exists {
					if strVal, ok := val.(string); ok {
						result = append(result, strVal)
						break
					}
				}
			}
		}
	}
	return result
}

func (c *Client) convertToStringList(items []interface{}) []string {
	result := []string{}
	for _, item := range items {
		if strVal, ok := item.(string); ok {
			result = append(result, strVal)
		} else {
			result = append(result, fmt.Sprintf("%v", item))
		}
	}
	return result
}
