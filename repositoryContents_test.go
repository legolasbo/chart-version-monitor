package main

import (
	"github.com/Masterminds/semver"
	"reflect"
	"testing"
)

func MapsEqual(a, b any, t *testing.T) {
	res := reflect.DeepEqual(a, b)
	if !res {
		t.Errorf("%v does dot equal %v", a, b)
	}
}

func TestRepositoryContents_FilterCharts(t *testing.T) {
	inputEntries := map[ChartName][]Entry{
		"remove":     {{Version: "1.0.0"}},
		"keep":       {{Version: "1.2.3"}},
		"remove_too": {{Version: "0.0.1"}},
	}

	rc := RepositoryContents{Entries: inputEntries}
	rc.FilterCharts([]Chart{{Name: "keep"}})

	Equals(len(rc.Entries), 1, t)
	Equals(rc.Entries["keep"][0], Entry{Version: "1.2.3"}, t)
}

func TestRepositoryContents_FilterCharts_NonExistingChart(t *testing.T) {
	inputEntries := map[ChartName][]Entry{
		"remove":     {{Version: "1.0.0"}},
		"keep":       {{Version: "1.2.3"}},
		"remove_too": {{Version: "0.0.1"}},
	}

	rc := RepositoryContents{Entries: inputEntries}
	rc.FilterCharts([]Chart{{Name: "keep_too"}})

	Equals(len(rc.Entries), 0, t)
}

func TestRepositoryContents_EntriesToVersions_NoEntries(t *testing.T) {
	rc := RepositoryContents{Entries: make(map[ChartName][]Entry)}

	rc.EntriesToVersions()

	MapsEqual(rc.Versions, make(map[ChartName]semver.Collection), t)
}

func TestRepositoryContents_EntriesToVersions_InvalidVersionString(t *testing.T) {
	rc := RepositoryContents{
		Entries: map[ChartName][]Entry{
			"chart": {
				Entry{Version: "Invalid"},
				Entry{Version: "1.0.0"},
			},
		},
	}

	rc.EntriesToVersions()

	initial, _ := semver.NewVersion("1.0.0")
	expected := map[ChartName]semver.Collection{
		"chart": {initial},
	}
	MapsEqual(rc.Versions, expected, t)
}

func TestRepositoryContents_EntriesToVersions(t *testing.T) {
	rc := RepositoryContents{
		Entries: map[ChartName][]Entry{
			"chart": {
				Entry{Version: "1.0.0"},
				Entry{Version: "1.0.1"},
				Entry{Version: "1.1.0"},
				Entry{Version: "2.0.0"},
			},
		},
	}

	rc.EntriesToVersions()

	initial, _ := semver.NewVersion("1.0.0")
	bugfix, _ := semver.NewVersion("1.0.1")
	minor, _ := semver.NewVersion("1.1.0")
	major, _ := semver.NewVersion("2.0.0")
	expected := map[ChartName]semver.Collection{
		"chart": {initial, bugfix, minor, major},
	}
	MapsEqual(rc.Versions, expected, t)
}

func TestRepositoryContents_EntriesToVersions_SortsVersions(t *testing.T) {
	rc := RepositoryContents{
		Entries: map[ChartName][]Entry{
			"chart": {
				Entry{Version: "2.0.0"},
				Entry{Version: "1.0.1"},
				Entry{Version: "1.0.0"},
				Entry{Version: "1.1.0"},
			},
		},
	}

	rc.EntriesToVersions()

	initial, _ := semver.NewVersion("1.0.0")
	bugfix, _ := semver.NewVersion("1.0.1")
	minor, _ := semver.NewVersion("1.1.0")
	major, _ := semver.NewVersion("2.0.0")
	expected := map[ChartName]semver.Collection{
		"chart": {initial, bugfix, minor, major},
	}
	MapsEqual(rc.Versions, expected, t)
}
