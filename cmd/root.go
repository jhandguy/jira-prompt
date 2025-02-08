package cmd

import (
	"fmt"
	"time"

	"github.com/jhandguy/jira-prompt/cmd/prompt"
	"github.com/jhandguy/jira-prompt/cmd/search"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var debug bool

var cmd = &cobra.Command{
	Use:   "jp",
	Short: "CLI to prompt Ollama with Jira data",
	Long:  "jira-prompt is a CLI to prompt Ollama using data from Jira issues.",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	cobra.OnInitialize(setup)

	cmd.AddCommand(search.Cmd)
	cmd.AddCommand(prompt.Cmd)

	cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug for jp")
	cmd.PersistentFlags().StringP("jira-url", "u", "https://ecosystem.atlassian.net", "jira base url")
	cmd.PersistentFlags().StringP("jira-token", "t", "", "jira API token")
	cmd.PersistentFlags().StringP("jira-request", "q", "{\"jql\": \"project = FRGE AND status = \\\"In Progress\\\"\", \"fields\": [\"summary\"]}", "jira search request")
	cmd.PersistentFlags().StringP("jira-excluded-fields", "e", "id,self,expand", "jira fields to exclude from the response (comma separated)")
}

func setup() {
	if err := setupLogger(); err != nil {
		fmt.Printf("failed to setup logger: %v", err)
	}
}

func setupLogger() error {
	config := zap.NewProductionConfig()
	config.Encoding = "console"
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.TimeOnly)
	config.DisableStacktrace = true
	config.DisableCaller = true

	if debug {
		config.Level.SetLevel(zap.DebugLevel)
	} else {
		config.Level.SetLevel(zap.InfoLevel)
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)
	return nil
}

func Execute(version string) {
	cmd.Version = version
	if err := cmd.Execute(); err != nil {
		zap.S().Fatalf("❌ %v", err)
	}
}
