# jira-prompt

**jira-prompt** is a CLI to prompt Ollama using data from Jira issues.

## Requirements

[Ollama](https://ollama.com/) must be installed and the server running with `ollama serve` to use **jira-prompt**.

## Installation

**jira-prompt** can be installed via [homebrew-tap](https://github.com/jhandguy/homebrew-tap) with

```shell
brew install jhandguy/tap/jira-prompt
```

or downloaded as binary from the [releases page](https://github.com/jhandguy/jira-prompt/releases).

## Usage

```shell
➜ ./jp
jira-prompt is a CLI to prompt Ollama using data from Jira issues.

Usage:
  jp [command]

Available Commands:
  help        Help about any command
  prompt      Prompt Ollama with Jira data
  search      Search Jira issues with JQL

Flags:
  -d, --debug                         debug for jp
  -h, --help                          help for jp
  -e, --jira-excluded-fields string   jira fields to exclude from the response (comma separated) (default "id,self,expand")
  -q, --jira-request string           jira search request (default "{\"jql\": \"project = FRGE AND status = \\\"In Progress\\\"\", \"fields\": [\"summary\"]}")
  -t, --jira-token string             jira API token
  -u, --jira-url string               jira base url (default "https://ecosystem.atlassian.net")
  -v, --version                       version for jp

Use "jp [command] --help" for more information about a command.
```
