package jira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRemoveKeys tests the removeKeys function in isolation.
func TestRemoveKeys(t *testing.T) {
	// Prepare a sample data map simulating a Jira response
	data := map[string]interface{}{
		"keep":   "value1",
		"remove": "value2",
		"nested": map[string]interface{}{
			"nestedRemove": "value3",
			"nestedKeep":   "value4",
		},
		"array": []interface{}{
			map[string]interface{}{
				"arrayRemove": "value5",
				"arrayKeep":   "value6",
			},
			"some-string-value",
		},
	}

	// Prepare unwanted keys
	unwantedKeys := map[string]bool{
		"remove":       true,
		"nestedRemove": true,
		"arrayRemove":  true,
	}

	// Call the function under test
	removeKeys(data, unwantedKeys)

	// Assert that "remove", "nestedRemove", and "arrayRemove" are removed
	// but that "keep", "nestedKeep", and "arrayKeep" remain.
	_, hasRemove := data["remove"]
	assert.False(t, hasRemove, "Expected 'remove' key to be deleted")

	nested, hasNested := data["nested"].(map[string]interface{})
	assert.True(t, hasNested, "Expected 'nested' map to exist")

	if hasNested {
		_, hasNestedRemove := nested["nestedRemove"]
		assert.False(t, hasNestedRemove, "Expected 'nestedRemove' key to be deleted")
		_, hasNestedKeep := nested["nestedKeep"]
		assert.True(t, hasNestedKeep, "Expected 'nestedKeep' key to remain")
	}

	arr, hasArray := data["array"].([]interface{})
	assert.True(t, hasArray, "Expected 'array' to exist")

	if hasArray && len(arr) > 0 {
		// First item is a nested map
		firstItem, ok := arr[0].(map[string]interface{})
		assert.True(t, ok, "Expected first item in array to be a map")

		if ok {
			_, hasArrayRemove := firstItem["arrayRemove"]
			assert.False(t, hasArrayRemove, "Expected 'arrayRemove' key to be deleted")
			_, hasArrayKeep := firstItem["arrayKeep"]
			assert.True(t, hasArrayKeep, "Expected 'arrayKeep' key to remain")
		}
	}

	// Ensure other keys/values remain untouched
	val, hasKeep := data["keep"]
	assert.True(t, hasKeep, "Expected 'keep' key to remain")
	assert.Equal(t, "value1", val, "Expected 'keep' key to be 'value1'")
}

// TestSearch_Success tests the Search method in a successful scenario (200 OK).
func TestSearch_Success(t *testing.T) {
	// Create a mock server to simulate Jira's /rest/api/2/search/jql endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the endpoint is correct
		assert.Equal(t, "/rest/api/2/search/jql", r.URL.Path)

		// Return a 200 status and some JSON
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
            "expand": "schema,names",
            "startAt": 0,
            "maxResults": 50,
            "total": 1,
            "issues": [
                {
                    "id": "10000",
                    "key": "PROJ-1",
                    "fields": {
                        "summary": "Issue summary"
                    }
                }
            ]
        }`)
	}))
	defer mockServer.Close()

	// Create a Jira instance that points to our mock server
	j := New(mockServer.URL, "test-auth-token")

	// Call the Search method with a filter that removes "expand"
	filters := "expand"
	result, err := j.Search(`{"jql":"project=PROJ"}`, filters)
	assert.NoError(t, err, "Expected no error on successful response")

	// Parse the resulting JSON
	var resultData map[string]interface{}
	err = json.Unmarshal([]byte(result), &resultData)
	assert.NoError(t, err, "Result should be valid JSON")

	// "expand" should have been removed due to the filter
	_, hasExpand := resultData["expand"]
	assert.False(t, hasExpand, "Expected 'expand' to be removed")

	// "issues" should remain, check if the length is 1
	issues, ok := resultData["issues"].([]interface{})
	assert.True(t, ok, "Expected 'issues' key to be an array")
	assert.Len(t, issues, 1, "Expected 1 issue in the array")

	// Optionally, verify other keys remain or are structured as expected
	_, hasStartAt := resultData["startAt"]
	assert.True(t, hasStartAt, "Expected 'startAt' key to remain")
}

// TestSearch_NonOK tests the Search method when Jira returns a non-200 response.
func TestSearch_NonOK(t *testing.T) {
	// Create a mock server that returns 400 for demonstration
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"error":"Bad request"}`)
	}))
	defer mockServer.Close()

	// Create a Jira instance that points to our mock server
	j := New(mockServer.URL, "test-auth-token")

	// Call the Search method, expecting an error
	_, err := j.Search(`{"jql":"invalid"}`, "")
	assert.Error(t, err, "Expected an error for non-200 responses")
	assert.Contains(t, err.Error(), "failed to search for Jira issues", "Error message should contain hint")
}

// TestSearch_InvalidJSON tests when the Jira server returns invalid JSON.
func TestSearch_InvalidJSON(t *testing.T) {
	// Create a mock server returning invalid JSON
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return something that can't be unmarshaled into a map[string]interface{}
		fmt.Fprintln(w, `invalid-json`)
	}))
	defer mockServer.Close()

	// Create a Jira instance that points to our mock server
	j := New(mockServer.URL, "test-auth-token")

	// Call the Search method, expecting an unmarshal error
	_, err := j.Search(`{"jql":"project=PROJ"}`, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal Jira issues", "Should return unmarshal error")
}
