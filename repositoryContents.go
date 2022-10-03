package main

import (
	"github.com/Masterminds/semver"
	"log"
	"sort"
)

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
