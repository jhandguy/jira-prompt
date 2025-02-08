package jira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type Jira struct {
	restClient *resty.Client
}

func New(baseURL, authToken string) *Jira {
	return &Jira{
		restClient: resty.
			New().
			SetBaseURL(baseURL).
			SetAuthScheme("Basic").
			SetAuthToken(authToken).
			SetHeader("Content-Type", "application/json"),
	}
}

func (j *Jira) Search(body, excludedFields string) (string, error) {
	zap.S().Infof("🔍 Searching Jira issues...")
	res, err := j.restClient.R().
		SetBody(body).
		Post("/rest/api/2/search/jql")
	if err != nil {
		return "", err
	}

	if res.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("failed to search for Jira issues: %s", res.Status())
	}

	var data map[string]interface{}
	if err = json.Unmarshal([]byte(res.String()), &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal Jira issues: %w", err)
	}

	fieldsToRemove := make(map[string]bool)
	for _, str := range strings.Split(excludedFields, ",") {
		fieldsToRemove[str] = true
	}

	removeKeys(data, fieldsToRemove)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Jira issues: %w", err)
	}

	zap.S().Infof("✅ Search successful!")
	return string(jsonData), nil
}

func removeKeys(data map[string]interface{}, unwantedKeys map[string]bool) {
	for key := range data {
		// If key is in the map of keys to remove, delete it
		if unwantedKeys[key] {
			delete(data, key)
			continue
		}

		// If the value is a nested map, recurse into it
		if nestedMap, ok := data[key].(map[string]interface{}); ok {
			removeKeys(nestedMap, unwantedKeys)
		}

		// If the value is an array, iterate and check for nested maps
		if nestedArray, ok := data[key].([]interface{}); ok {
			for _, item := range nestedArray {
				if itemMap, ok := item.(map[string]interface{}); ok {
					removeKeys(itemMap, unwantedKeys)
				}
			}
		}
	}
}
