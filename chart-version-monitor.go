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
	"sort"
	"strings"
	"time"
)

func main() {
	config := getConfig()
	fixRepoURLS(config)
	log.Println(config)

	repositoriesToCheckForUpdates := make(chan *RepositoryContents)
	versionsToReport := make(chan Report)
	go checkRepositoriesForUpdates(repositoriesToCheckForUpdates, versionsToReport)
	go reportNewVersions(config, versionsToReport)

	ticker := time.NewTicker(config.CheckInterval.Duration())
	go sendStartInfo(config)
	go fetchAllRepositories(config, repositoriesToCheckForUpdates)
	for {
		select {
		case <-ticker.C:
			fetchAllRepositories(config, repositoriesToCheckForUpdates)
		}
	}
}

func sendStartInfo(config Config) {
	if !config.ReportStart {
		return
	}
	s := fmt.Sprintf("%s :: %s\n%s", time.Now().Format("2006-01-02 15:04:05"), "Helmchart monitor started", config)
	sendMessageToSlack(config, Message{Text: s})
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
		dependees := config.DependeesForChart(report.Repository, report.Chart)
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
	config := DefaultConfig().FromFile("config.json")

	PopulateRepositoriesFromEnvironment("CVM_REPOSITORIES", &config.Repositories)
	PopulateStringFromEnvironment("CVM_WEBHOOK_URL", &config.WebhookURL)
	PopulateBooleanFromEnvironment("CVM_REPORT_START", &config.ReportStart)
	PopulateDurationFromEnvironment("CVM_CHECK_INTERVAL", &config.CheckInterval)

	if config.Repositories == nil {
		log.Fatalln("No repositories configured")
	}

	if config.WebhookURL == "" {
		log.Fatalln("No webhookURL configured")
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
