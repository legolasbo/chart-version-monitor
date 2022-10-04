package main

import (
	"github.com/Masterminds/semver"
	"log"
	"sort"
)

type ChartName string

type RepositoryContents struct {
	Entries   map[ChartName][]Entry `yaml:"entries"`
	Versions  map[ChartName]semver.Collection
	Generated string `yaml:"generated"`
	URL       string
}

type Entry struct {
	Version string `yaml:"version"`
}

func (rc *RepositoryContents) FilterCharts(chartsToKeep []Chart) {
	filtered := make(map[ChartName][]Entry)

	for _, chart := range chartsToKeep {
		if entries, ok := rc.Entries[chart.Name]; ok {
			filtered[chart.Name] = entries
			continue
		}
	}

	rc.Entries = filtered
}

func (rc *RepositoryContents) EntriesToVersions() {
	if rc.Versions == nil {
		rc.Versions = make(map[ChartName]semver.Collection)
	}
	for k, e := range rc.Entries {
		versions := make(semver.Collection, 0)
		for _, v := range e {
			version, err := semver.NewVersion(v.Version)
			if err != nil {
				log.Printf("Error parsing version: %s", err)
				continue
			}

			versions = append(versions, version)
		}

		sort.Sort(versions)
		rc.Versions[k] = versions
	}
}
