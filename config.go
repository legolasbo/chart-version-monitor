package main

import (
	"encoding/json"
	"fmt"
)

type Repository struct {
	URL    string  `json:"url"`
	Charts []Chart `json:"charts"`
}

type Chart struct {
	Name      string   `json:"name"`
	Dependees []string `json:"dependees"`
}

type Config struct {
	Repositories  []Repository `json:"repositories"`
	CheckInterval string       `json:"checkInterval"`
	WebhookURL    string       `json:"webhookURL"`
	ReportStart   bool         `json:"reportStart"`
}

func (c Config) String() string {
	repositories, _ := json.MarshalIndent(c.Repositories, "", "  ")
	return fmt.Sprintf(`Configuration:
Webhook: %s
Check interval: %s
Report start: %t
Repositories:
%s
`, c.WebhookURL, c.CheckInterval, c.ReportStart, "```\n"+string(repositories)+"\n```")
}

func (c *Config) DependeesForChart(repository string, chart string) []string {
	for _, c := range c.ChartsForRepository(repository) {
		if c.Name == chart {
			return c.Dependees
		}
	}

	return make([]string, 0)
}

func (c *Config) ChartsForRepository(repository string) []Chart {
	for _, r := range c.Repositories {
		if r.URL == repository {
			return r.Charts
		}
	}

	return make([]Chart, 0)
}
