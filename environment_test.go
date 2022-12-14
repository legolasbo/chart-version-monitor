package main

import (
	"os"
	"testing"
	"time"
)

const NonExisting = "NON_EXISTING_ENVIRONMENT_VARIABLE"
const Existing = "EXISTING_VARIABLE"

func Equals(a interface{}, b interface{}, t *testing.T) {
	if a != b {
		t.Errorf("%v does not equals %v", a, b)
	}
}

func TestPopulateBooleanFromEnvironment_NoVariablePresent(t *testing.T) {
	test := true
	success := PopulateBooleanFromEnvironment(NonExisting, &test)

	Equals(test, true, t)
	Equals(success, false, t)

	test = false
	success = PopulateBooleanFromEnvironment(NonExisting, &test)

	Equals(test, false, t)
	Equals(success, false, t)
}

func TestPopulateBooleanFromEnvironment_VariableUnparsable(t *testing.T) {
	_ = os.Setenv(Existing, "unparsable")
	test := true
	success := PopulateBooleanFromEnvironment(Existing, &test)

	Equals(test, true, t)
	Equals(success, false, t)

	test = false
	success = PopulateBooleanFromEnvironment(Existing, &test)

	Equals(test, false, t)
	Equals(success, false, t)
}

func TestPopulateBooleanFromEnvironment(t *testing.T) {
	_ = os.Setenv(Existing, "t")
	test := false
	success := PopulateBooleanFromEnvironment(Existing, &test)

	Equals(test, true, t)
	Equals(success, true, t)
}

func TestPopulateStringFromEnvironment_NonExisting(t *testing.T) {
	test := "startValue"
	success := PopulateStringFromEnvironment(NonExisting, &test)

	Equals(test, "startValue", t)
	Equals(success, false, t)
}

func TestPopulateStringFromEnvironment(t *testing.T) {
	_ = os.Setenv(Existing, "newValue")
	test := "startValue"
	success := PopulateStringFromEnvironment(Existing, &test)

	Equals(test, "newValue", t)
	Equals(success, true, t)
}

func TestPopulateRepositoriesFromEnvironment_NonExisting(t *testing.T) {
	test := make([]Repository, 0)
	success := PopulateRepositoriesFromEnvironment(NonExisting, &test)

	Equals(len(test), 0, t)
	Equals(success, false, t)
}

func TestPopulateRepositoriesFromEnvironment_NonParsable(t *testing.T) {
	_ = os.Setenv(Existing, "blaat")
	test := make([]Repository, 0)
	success := PopulateRepositoriesFromEnvironment(Existing, &test)

	Equals(len(test), 0, t)
	Equals(success, false, t)
}

func TestPopulateRepositoriesFromEnvironment(t *testing.T) {
	fakeRepository := `[
    {
      "url": "http://localhost:8080/fakechart.yaml",
      "charts": [
        {
          "name": "test",
          "dependees": [
            "Onegini",
            "Boerenkool"
          ]
        }
      ]
    }
  ]`

	_ = os.Setenv(Existing, fakeRepository)

	test := make([]Repository, 0)
	success := PopulateRepositoriesFromEnvironment(Existing, &test)

	Equals(len(test), 1, t)
	Equals(success, true, t)

	firstRepo := test[0]
	Equals(firstRepo.URL, "http://localhost:8080/fakechart.yaml", t)
	Equals(len(firstRepo.Charts), 1, t)
}

func TestPopulateDurationFromEnvironment_NonExisting(t *testing.T) {
	d := Duration(100)

	success := PopulateDurationFromEnvironment(NonExisting, &d)

	Equals(d, Duration(100), t)
	Equals(success, false, t)
}

func TestPopulateDurationFromEnvironment_unparsable(t *testing.T) {
	_ = os.Setenv(Existing, "fails")
	d := Duration(100)

	success := PopulateDurationFromEnvironment(Existing, &d)

	Equals(d, Duration(100), t)
	Equals(success, false, t)
}

func TestPopulateDurationFromEnvironment(t *testing.T) {
	_ = os.Setenv(Existing, "10m")
	d := Duration(10 * time.Minute)

	success := PopulateDurationFromEnvironment(Existing, &d)

	Equals(d, Duration(10*time.Minute), t)
	Equals(success, true, t)
}
