package main

import "encoding/json"

type BuildParameters struct {
	Branch      string `json:"branch"`
	Version     string `json:"version"`
	BuildNumber string `json:"build_number"`

	CurrentVersion     string `json:"current_version"`
	CurrentBuildNumber string `json:"current_build_number"`
	NextPatch          string `json:"next_patch"`
	NextMinor          string `json:"next_minor"`
	NextMajor          string `json:"next_major"`
	NextBuildNumber    string `json:"next_build_number"`
	InfoPlist          string `json:"info_plist"`
}

func NewBuildParameters(jsonStr string) BuildParameters {
	parameters := new(BuildParameters)
	json.Unmarshal([]byte(jsonStr), parameters)
	return *parameters
}

func (v BuildParameters) string() string {
	bytes, _ := json.Marshal(v)
	return string(bytes[:])
}
