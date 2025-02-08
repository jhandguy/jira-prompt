package search

import (
	"fmt"

	"github.com/jhandguy/jira-prompt/internal/jira"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:           "search",
	Short:         "Search Jira issues with JQL",
	RunE:          search,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func search(cmd *cobra.Command, _ []string) error {
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

	res, err := jira.
		New(jiraURL, jiraToken).
		Search(jiraRequest, jiraExcludedFields)
	if err != nil {
		return err
	}

	fmt.Print(res)
	return nil
}
