package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
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

func DefaultConfig() Config {
	return Config{
		CheckInterval: "1h",
		ReportStart:   true,
	}
}

func (c Config) FromFile(name string) Config {
	configFile, err := os.Open(name)
	if err != nil {
		return c
	}
	defer configFile.Close()

	log.Println("Successfully opened", name)
	configBytes, _ := io.ReadAll(configFile)

	config := c
	err = json.Unmarshal(configBytes, c)
	if err != nil {
		return c
	}
	return config
}

func (c Config) FromEnvironment() Config {
	return c
}
