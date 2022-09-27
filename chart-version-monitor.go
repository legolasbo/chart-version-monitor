package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

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

func (c *Config) dependeesForChart(repository string, chart string) []string {
	for _, c := range c.chartsForRepository(repository) {
		if c.Name == chart {
			return c.Dependees
		}
	}

	return make([]string, 0)
}

func (c *Config) chartsForRepository(repository string) []Chart {
	for _, r := range c.Repositories {
		if r.URL == repository {
			return r.Charts
		}
	}

	return make([]Chart, 0)
}

type Repository struct {
	URL    string  `json:"url"`
	Charts []Chart `json:"charts"`
}

type Chart struct {
	Name      string   `json:"name"`
	Dependees []string `json:"dependees"`
}

type RepositoryContents struct {
	Entries   map[string][]Entry `yaml:"entries"`
	Versions  map[string]semver.Collection
	Generated string `yaml:"generated"`
	URL       string
}

type Entry struct {
	Version string `yaml:"version"`
}

func (rc *RepositoryContents) filterCharts(chartsToKeep []Chart) {
	filtered := make(map[string][]Entry)

	for _, chart := range chartsToKeep {
		if entries, ok := rc.Entries[chart.Name]; ok {
			filtered[chart.Name] = entries
			continue
		}
		rc.Entries[chart.Name] = make([]Entry, 0)
	}

	rc.Entries = filtered
}

func (rc *RepositoryContents) entriesToVersions() {
	if rc.Versions == nil {
		rc.Versions = make(map[string]semver.Collection)
	}
	for k, e := range rc.Entries {
		versions := make(semver.Collection, len(e))
		for i, v := range e {
			version, err := semver.NewVersion(v.Version)
			if err != nil {
				log.Printf("Error parsing version: %s", err)
				continue
			}

			versions[i] = version
		}

		sort.Sort(versions)
		rc.Versions[k] = versions
	}
}

func main() {
	config := getConfig()
	fixRepoURLS(config)
	log.Println(config)

	repositoriesToCheckForUpdates := make(chan *RepositoryContents)
	versionsToReport := make(chan Report)
	startInfo := make(chan string, 3)
	go checkRepositoriesForUpdates(repositoriesToCheckForUpdates, versionsToReport)
	go reportNewVersions(config, versionsToReport)

	interval, err := time.ParseDuration(config.CheckInterval)
	if err != nil {
		log.Fatalln("Invalid checkInterval. Must be a valid Golang duration string such as 10s, 1m10s or 1h20m30s")
	}
	ticker := time.NewTicker(interval)
	startInfo <- fmt.Sprintf("%s :: %s", time.Now().Format("2006-01-02 15:04:05"), "Helmchart monitor started")
	startInfo <- config.String()
	for {
		select {
		case s := <-startInfo:
			if config.ReportStart {
				sendMessageToSlack(config, Message{Text: s})
			}
		case <-ticker.C:
			fetchAllRepositories(config, repositoriesToCheckForUpdates)
		}
	}
}

func fetchAllRepositories(config Config, repositoriesToCheckForUpdates chan *RepositoryContents) {
	for _, repo := range config.Repositories {
		repoContents, err := fetchRepositoryContents(repo)
		if err != nil {
			log.Println("Could not fetch contents for", repo.URL, err)
			continue
		}

		repositoriesToCheckForUpdates <- repoContents
	}
}

type Report struct {
	Repository, Chart string
	NewVersion        *semver.Version
}

func checkRepositoriesForUpdates(toCheck <-chan *RepositoryContents, toReport chan<- Report) {
	highestVersions := make(map[string]map[string]*semver.Version)
	for repo := range toCheck {
		log.Println("Checking:", repo.URL)
		if highestVersions[repo.URL] == nil {
			highestVersions[repo.URL] = make(map[string]*semver.Version)
		}

		for chartName, versions := range repo.Versions {
			sort.Sort(versions)
			highestVersion := versions[len(versions)-1]

			currentVersion, ok := highestVersions[repo.URL][chartName]
			if !ok {
				highestVersions[repo.URL][chartName] = highestVersion
				continue
			}

			if currentVersion.LessThan(highestVersion) {
				highestVersions[repo.URL][chartName] = highestVersion
				toReport <- Report{
					Repository: repo.URL,
					Chart:      chartName,
					NewVersion: highestVersion,
				}
			}
		}

	}
}

type Message struct {
	Text string `json:"text"`
}

func reportNewVersions(config Config, toReport <-chan Report) {
	for report := range toReport {
		msg := Message{
			Text: fmt.Sprintf("Chart *%s* in repo %s updated to version *%s*", report.Chart, report.Repository, report.NewVersion),
		}
		dependees := config.dependeesForChart(report.Repository, report.Chart)
		if len(dependees) > 0 {
			msg.Text = fmt.Sprintln(msg.Text, "\nYou might want to check:", strings.Join(dependees, ", "))
		}

		sendMessageToSlack(config, msg)
		log.Println(report)
	}
}

func sendMessageToSlack(config Config, msg Message) {
	data, _ := json.Marshal(msg)
	response, err := http.Post(config.WebhookURL, http.DetectContentType(data), bytes.NewReader(data))
	if err != nil {
		log.Println(err)
	}
	if response.StatusCode > 300 {
		log.Println("Weird response code whilst reporting", response.StatusCode)
	}
}

func fetchRepositoryContents(repo Repository) (*RepositoryContents, error) {
	resp, err := http.Get(repo.URL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 400 {
		return nil, errors.New("Unable to fetch: " + repo.URL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	repoContents := RepositoryContents{URL: repo.URL}
	err = yaml.Unmarshal(body, &repoContents)
	if err != nil {
		return nil, err
	}

	repoContents.filterCharts(repo.Charts)
	repoContents.entriesToVersions()
	return &repoContents, nil
}

func getConfig() Config {
	config := readConfigFile()

	value, present := os.LookupEnv("CVM_REPOSITORIES")
	if present {
		repos := make([]Repository, 0)
		err := json.Unmarshal([]byte(value), &repos)
		if err == nil {
			config.Repositories = repos
		}
	}

	value, present = os.LookupEnv("CVM_WEBHOOK_URL")
	if present {
		config.WebhookURL = value
	}

	value, present = os.LookupEnv("CVM_REPORT_START")
	if present {
		v, err := strconv.ParseBool(value)
		if err == nil {
			config.ReportStart = v
		}
	}

	value, present = os.LookupEnv("CVM_CHECK_INTERVAL")
	if present {
		_, err := time.ParseDuration(value)
		if err == nil {
			config.CheckInterval = value
		}
	}

	if config.Repositories == nil {
		log.Fatalln("No repositories configured")
	}

	if config.WebhookURL == "" {
		log.Fatalln("No webhookURL configured")
	}
	return config
}

func readConfigFile() Config {
	defaultConfig := Config{
		CheckInterval: "1h",
		ReportStart:   true,
	}

	configFile, err := os.Open("config.json")
	if err != nil {
		return defaultConfig
	}
	defer configFile.Close()

	fmt.Println("Successfully opened config.json")
	configBytes, _ := io.ReadAll(configFile)

	config := defaultConfig
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return defaultConfig
	}
	return config
}

func fixRepoURLS(config Config) {
	for i, repo := range config.Repositories {
		repoIndex := "/index.yaml"
		extension := ".yaml"
		if repo.URL[len(repo.URL)-len(extension):] != extension {
			repo.URL += repoIndex
			config.Repositories[i] = repo
		}
	}
}
