package ollama

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type Ollama struct {
	restClient *resty.Client
}

type request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Raw    bool   `json:"raw"`
}

func New(baseURL string) *Ollama {
	return &Ollama{
		restClient: resty.
			New().
			SetBaseURL(baseURL).
			SetHeader("Content-Type", "application/json"),
	}
}

func (o *Ollama) Prompt(model, textPrompt, jiraResponse string, stream, raw bool) (string, error) {
	prompt := fmt.Sprintf("%s\n%s", textPrompt, jiraResponse)
	zap.S().Infof("💬 Prompting %s model...", model)
	zap.S().Debug(prompt)

	res, err := o.restClient.R().
		SetDoNotParseResponse(stream).
		SetBody(&request{
			Model:  model,
			Prompt: prompt,
			Stream: stream,
			Raw:    raw,
		}).
		Post("/api/generate")
	if err != nil {
		return "", err
	}
	defer res.RawBody().Close()

	if res.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("failed to prompt %s: %s", model, res.Status())
	}

	zap.S().Infof("✅ Prompt successful!")

	if !stream {
		return unmarshallResponse([]byte(res.String()))
	}

	buf := make([]byte, 10240)
	for {
		n, readErr := res.RawBody().Read(buf)
		if readErr == io.EOF {
			// We've read all the data
			break
		}
		if readErr != nil {
			log.Fatalf("error reading from stream: %v", readErr)
		}

		// Process the chunk that was read
		chunk := buf[:n]
		resp, err := unmarshallResponse(chunk)
		if err != nil {
			return "", err
		}

		fmt.Print(resp)
	}

	return "", nil
}

func unmarshallResponse(response []byte) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(response, &data); err != nil {
		return "", fmt.Errorf("failed unmarshal generated response: %w", err)
	}

	resp, ok := data["response"].(string)
	if !ok {
		return "", fmt.Errorf("the \"response\" field is missing or not a string in the returned JSON: %v", data)
	}

	return resp, nil
}
