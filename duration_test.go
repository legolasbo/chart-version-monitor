package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDuration_MarshalJSON(t *testing.T) {
	d, _ := time.ParseDuration("10s")
	duration := Duration(d)

	result, err := duration.MarshalJSON()
	if err != nil {
		t.Errorf("Should not fail")
	}

	Equals(string(result), "\"10s\"", t)
}

func TestDuration_UnmarshalJSON_float(t *testing.T) {
	var d Duration

	val := float64(1.123)
	j, _ := json.Marshal(val)
	err := d.UnmarshalJSON(j)

	Equals(err, nil, t)
	Equals(d, Duration(time.Duration(val)), t)
}

func TestDuration_UnmarshallJSON_int(t *testing.T) {
	var d Duration

	j, _ := json.Marshal(10)
	err := d.UnmarshalJSON(j)

	Equals(err, nil, t)
	Equals(d, Duration(time.Duration(10)), t)
}

func TestDuration_UnmarshallJSON_parsableString(t *testing.T) {
	initialDuration := Duration(1000000)
	j, _ := json.Marshal(initialDuration)

	Equals(string(j), "\"1ms\"", t)

	var d Duration
	err := d.UnmarshalJSON(j)

	Equals(err, nil, t)
	Equals(d, initialDuration, t)
}

func TestParseDuration_parseError(t *testing.T) {
	d, err := ParseDuration("error")

	var emptyDuration Duration
	Equals(d, emptyDuration, t)
	Equals(err != nil, true, t)
}

func TestParseDuration(t *testing.T) {
	d, err := ParseDuration("1h10m50s")

	Equals(err, nil, t)
	Equals(d.String(), "1h10m50s", t)
}
