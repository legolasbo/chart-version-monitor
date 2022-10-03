package main

import (
	"encoding/json"
	"os"
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
	err := json.Unmarshal([]byte(value), &repos)
	if err != nil {
		return false
	}

	*variable = repos
	return true
}
