package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrompt_Success(t *testing.T) {
	// Create a mock server to simulate /api/generate endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method and path
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/generate", r.URL.Path)

		// Decode the JSON body from the request to verify the payload
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err, "Expected to decode request body without error")

		// Check that the request fields match what we expect
		assert.Equal(t, "test-model", reqBody["model"], "model should match")
		// We expect the prompt to contain both textPrompt and jiraResponse separated by newline
		expectedPrompt := "Hello from text\nSome JIRA data"
		assert.Equal(t, expectedPrompt, reqBody["prompt"], "prompt should match combined input")
		assert.Equal(t, false, reqBody["stream"], "stream should be false")

		// Return a 200 response with a JSON body
		w.WriteHeader(http.StatusOK)
		// The response is expected to contain a "response" field
		fmt.Fprintf(w, `{"response":"Model output response"}`)
	}))
	defer mockServer.Close()

	// Create an Ollama client pointing to our mock server
	o := New(mockServer.URL)

	// Call Prompt with some sample parameters
	model := "test-model"
	textPrompt := "Hello from text"
	jiraResponse := "Some JIRA data"
	resp, err := o.Prompt(model, textPrompt, jiraResponse, false, false)

	// Assertions
	assert.NoError(t, err, "Expected no error on successful call")
	assert.Equal(t, "Model output response", resp, "Should match the response from mock server")
}

func TestPrompt_NonOK(t *testing.T) {
	// Mock server returns a non-200 status code
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"Bad Request"}`)
	}))
	defer mockServer.Close()

	o := New(mockServer.URL)

	_, err := o.Prompt("failing-model", "any prompt", "any Jira data", false, false)
	assert.Error(t, err, "Expected error for non-200 response")
	assert.Contains(t, err.Error(), "failed to prompt failing-model: 400 Bad Request")
}

func TestPrompt_InvalidJSON(t *testing.T) {
	// Mock server returns a 200, but invalid JSON
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `invalid-json...`)
	}))
	defer mockServer.Close()

	o := New(mockServer.URL)

	_, err := o.Prompt("test-model", "text prompt", "jira data", false, false)
	assert.Error(t, err, "Expected JSON unmarshal error")
	assert.Contains(t, err.Error(), "failed unmarshal generated response", "Error message should mention unmarshal")
}

func TestPrompt_ResponseFieldIsMissingOrWrongType(t *testing.T) {
	// We'll run two sub-tests (table-driven) to cover:
	//   1. The "response" field is missing.
	//   2. The "response" field exists but is not a string.
	testCases := []struct {
		name         string
		responseBody string
	}{
		{
			name:         "missingResponseField",
			responseBody: `{"notresponse": "hello"}`, // "response" field is absent
		},
		{
			name:         "wrongTypeResponseField",
			responseBody: `{"response": 12345}`, // "response" is present but not a string
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Return 200 OK with a malformed body
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, tc.responseBody)
			}))
			defer mockServer.Close()

			// Use the Ollama client pointing to the mock server
			o := New(mockServer.URL)

			// Call Prompt; expect an error about the "response" field
			_, err := o.Prompt("test-model", "prompt", "jira data", false, false)
			assert.Error(t, err, "Expected an error due to a missing or invalid 'response' field")
			assert.Contains(t, err.Error(), `the "response" field is missing or not a string`,
				"Error message should mention the missing/invalid response field")
		})
	}
}

func TestPrompt_Stream(t *testing.T) {
	tests := []struct {
		name       string
		raw        bool
		chunkParts []string // Each part will be written as its own JSON object + newline
	}{
		{
			name: "StreamRawFalse",
			raw:  false,
			chunkParts: []string{
				`{"response":"Chunk #1 - raw=false"}`,
				`{"response":"Chunk #2 - raw=false"}`,
			},
		},
		{
			name: "StreamRawTrue",
			raw:  true,
			chunkParts: []string{
				`{"response":"Chunk #1 - raw=true"}`,
				`{"response":"Chunk #2 - raw=true"}`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create an HTTP test server
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Ensure we can flush
				flusher, ok := w.(http.Flusher)
				if !ok {
					t.Fatal("expected http.ResponseWriter to be an http.Flusher")
				}

				// Parse the request body to confirm the "stream" and "raw" fields
				var reqBody map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				assert.NoError(t, err)

				assert.Equal(t, true, reqBody["stream"], "Expected stream=true for streaming scenario")
				assert.Equal(t, tc.raw, reqBody["raw"], "Raw flag in request body should match test case")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				// Write each chunk on its own line, then flush (and optionally sleep).
				// This helps ensure each chunk arrives (and is read) separately.
				for i, part := range tc.chunkParts {
					// Write JSON + newline
					fmt.Fprintln(w, part)
					flusher.Flush()

					// (Optional) Sleep to reduce coalescing
					time.Sleep(50 * time.Millisecond)

					_ = i
				}
			}))
			defer mockServer.Close()

			// Create Ollama client pointing to the mock server
			client := New(mockServer.URL)

			// Capture stdout, since streaming prints directly to stdout
			oldStdout := os.Stdout
			r, wPipe, _ := os.Pipe()
			os.Stdout = wPipe

			// Call the Prompt method in streaming mode
			result, err := client.Prompt("stream-model", "stream prompt", "jira data", true, tc.raw)

			// Close and restore stdout
			wPipe.Close()
			os.Stdout = oldStdout

			// Read the captured output
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			// In streaming mode, we expect an empty return string
			assert.NoError(t, err)
			assert.Empty(t, result, "Expected empty result when streaming")

			// Each chunk's "response" value should appear in the captured output
			for _, chunk := range tc.chunkParts {
				var data map[string]interface{}
				err := json.Unmarshal([]byte(chunk), &data)
				assert.NoError(t, err)

				resp, ok := data["response"].(string)
				assert.True(t, ok, "test chunk must have a string 'response' field")
				assert.Contains(t, output, resp,
					"Expected printed output to contain chunk's 'response' value.")
			}
		})
	}
}
