package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"sigs.k8s.io/yaml"
)

const ENV_Repositories = "CVM_REPOSITORIES"
const ENV_WebhookURL = "CVM_WEBHOOK_URL"
const ENV_ReportStart = "CVM_REPORT_START"
const ENV_CheckInterval = "CVM_CHECK_INTERVAL"

type Repository struct {
	URL    string  `json:"url"`
	Charts []Chart `json:"charts"`
}

func (r Repository) Validate() error {
	if r.URL == "" {
		return errors.New("the repository URL should not be empty")
	}

	if len(r.Charts) == 0 {
		return fmt.Errorf("repository %s is not configured to monitor any charts", r.URL)
	}

	return nil
}

type Chart struct {
	Name      ChartName `json:"name"`
	Dependees []string  `json:"dependees"`
}

type Config struct {
	Repositories  []Repository `json:"repositories"`
	CheckInterval Duration     `json:"check_interval"`
	WebhookURL    string       `json:"webhook_url"`
	ReportStart   bool         `json:"report_start"`
}

func (c Config) String() string {
	repositories, _ := yaml.Marshal(c.Repositories)
	return fmt.Sprintf(`Configuration:
Webhook: %s
Check interval: %s
Report start: %t
Repositories:
%s
`, c.WebhookURL, c.CheckInterval, c.ReportStart, "```\n"+string(repositories)+"```")
}

func (c *Config) DependeesForChart(repository string, chart ChartName) []string {
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
		CheckInterval: Duration(1 * time.Hour),
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

	originalConfig := c
	err = yaml.Unmarshal(configBytes, &c)
	if err != nil {
		log.Println("Could not open", name, err)
		return originalConfig
	}
	return c
}

func (c Config) FromEnvironment() Config {
	PopulateRepositoriesFromEnvironment(ENV_Repositories, &c.Repositories)
	PopulateStringFromEnvironment(ENV_WebhookURL, &c.WebhookURL)
	PopulateBooleanFromEnvironment(ENV_ReportStart, &c.ReportStart)
	PopulateDurationFromEnvironment(ENV_CheckInterval, &c.CheckInterval)
	return c
}

func (c Config) Validate() error {
	if c.Repositories == nil {
		return errors.New("no repositories configured")
	}

	if len(c.Repositories) == 0 {
		return errors.New("no repositories configured")
	}

	for _, r := range c.Repositories {
		err := r.Validate()
		if err != nil {
			return fmt.Errorf("repositories contains an invalid repository: %w", err)
		}
	}

	if c.WebhookURL == "" {
		return errors.New("no webhookURL configured")
	}

	return nil
}
