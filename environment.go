package main

import (
	"os"
	"sigs.k8s.io/yaml"
	"strconv"
)

func PopulateBooleanFromEnvironment(name string, variable *bool) bool {
	value, present := os.LookupEnv(name)
	if !present {
		return false
	}

	v, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}

	*variable = v
	return true
}

func PopulateStringFromEnvironment(name string, variable *string) bool {
	value, present := os.LookupEnv(name)
	if !present {
		return false
	}

	*variable = value
	return true
}

func PopulateRepositoriesFromEnvironment(name string, variable *[]Repository) bool {
	value, present := os.LookupEnv(name)
	if !present {
		return false
	}

	repos := make([]Repository, 0)
	err := yaml.Unmarshal([]byte(value), &repos)
	if err != nil {
		return false
	}

	*variable = repos
	return true
}

func PopulateDurationFromEnvironment(name string, variable *Duration) bool {
	value, present := os.LookupEnv(name)
	if !present {
		return false
	}

	duration, err := ParseDuration(value)
	if err != nil {
		return false
	}

	*variable = duration
	return true
}
