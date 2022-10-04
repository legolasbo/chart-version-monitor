package main

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func ErrorsEqual(a, b error, t *testing.T) {
	Equals(a.Error(), b.Error(), t)
}

func TestConfig_Validate_NoRepositories(t *testing.T) {
	c := Config{}

	err := c.Validate()

	ErrorsEqual(err, errors.New("no repositories configured"), t)
}

func TestConfig_Validate_EmptyRepositories(t *testing.T) {
	c := Config{
		Repositories: make([]Repository, 0),
	}

	err := c.Validate()

	ErrorsEqual(err, errors.New("no repositories configured"), t)
}

func TestConfig_Validate_InvalidRepository(t *testing.T) {
	c := Config{
		Repositories: []Repository{
			{
				URL:    "",
				Charts: nil,
			},
		},
	}

	err := c.Validate()

	ErrorsEqual(err, fmt.Errorf("repositories contains an invalid repository: %w", errors.New("the repository URL should not be empty")), t)
}

func TestConfig_Validate_NoWebhookUrl(t *testing.T) {
	c := Config{
		Repositories: []Repository{
			{
				URL: "https://example.com",
				Charts: []Chart{
					{},
				},
			},
		},
	}

	err := c.Validate()

	ErrorsEqual(err, errors.New("no webhookURL configured"), t)
}

func TestConfig_Validate(t *testing.T) {
	c := Config{
		Repositories: []Repository{
			{
				URL: "https://example.com",
				Charts: []Chart{
					{},
				},
			},
		},
		WebhookURL: "https://example.com",
	}

	err := c.Validate()

	Equals(err, nil, t)
}

func TestConfig_FromEnvironment(t *testing.T) {
	_ = os.Setenv(ENV_Repositories, "[{\"url\": \"https://example.com/index.yaml\",\"charts\": [{\"name\": \"test\",\"dependees\": [\"some\"]}]}]")
	_ = os.Setenv(ENV_WebhookURL, "https://example.com/web/hook")
	_ = os.Setenv(ENV_ReportStart, "true")
	_ = os.Setenv(ENV_CheckInterval, "1h")

	c := DefaultConfig().FromEnvironment()

	Equals(len(c.Repositories), 1, t)
	Equals(c.WebhookURL, "https://example.com/web/hook", t)
	Equals(c.ReportStart, true, t)
	Equals(c.CheckInterval, Duration(1*time.Hour), t)
}
