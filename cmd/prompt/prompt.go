package prompt

import (
	"fmt"

	"github.com/jhandguy/jira-prompt/internal/jira"
	"github.com/jhandguy/jira-prompt/internal/ollama"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:           "prompt",
	Short:         "Prompt Ollama with Jira data",
	RunE:          prompt,
	SilenceUsage:  true,
	SilenceErrors: true,
}

var (
	ollamaHost, ollamaModel, ollamaPrompt string
	ollamaStream, ollamaRaw               bool
)

func init() {
	Cmd.Flags().StringVarP(&ollamaHost, "ollama-host", "o", "http://127.0.0.1:11434", "ollama host url")
	Cmd.Flags().StringVarP(&ollamaModel, "ollama-model", "m", "llama3", "ollama AI model")
	Cmd.Flags().StringVarP(&ollamaPrompt, "ollama-prompt", "p", "Given the following JSON representation of a Jira board, describe what the Forge team is working on:", "ollama text prompt")
	Cmd.Flags().BoolVarP(&ollamaStream, "ollama-stream", "s", true, "enable ollama streaming")
	Cmd.Flags().BoolVarP(&ollamaRaw, "ollama-raw", "r", false, "disable ollama formatting")
}

func prompt(cmd *cobra.Command, _ []string) error {
	jiraURL, err := cmd.InheritedFlags().GetString("jira-url")
	if err != nil {
		return err
	}

	jiraToken, err := cmd.InheritedFlags().GetString("jira-token")
	if err != nil {
		return err
	}

	jiraRequest, err := cmd.InheritedFlags().GetString("jira-request")
	if err != nil {
		return err
	}

	jiraExcludedFields, err := cmd.InheritedFlags().GetString("jira-excluded-fields")
	if err != nil {
		return err
	}

	jiraResponse, err := jira.
		New(jiraURL, jiraToken).
		Search(jiraRequest, jiraExcludedFields)
	if err != nil {
		return err
	}

	res, err := ollama.
		New(ollamaHost).
		Prompt(ollamaModel, ollamaPrompt, jiraResponse, ollamaStream, ollamaRaw)
	if err != nil {
		return err
	}

	fmt.Print(res)
	return nil
}
