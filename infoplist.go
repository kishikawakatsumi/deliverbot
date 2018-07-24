package main

import (
	"github.com/blang/semver"
	"howett.net/plist"
	"io/ioutil"
	"strconv"
)

const (
	versionKey     = "CFBundleShortVersionString"
	buildNumberKey = "CFBundleVersion"
)

type InfoPlist struct {
	Path   string
	object map[string]interface{}
	raw    []byte
}

func NewInfoPlist(path string) (*InfoPlist, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var object map[string]interface{}
	if _, err := plist.Unmarshal(bytes, &object); err != nil {
		return nil, err
	}

	infoPlist := InfoPlist{
		Path:   path,
		object: object,
		raw:    bytes,
	}

	return &infoPlist, nil
}

func (infoPlist *InfoPlist) VersionString() string {
	return infoPlist.object[versionKey].(string)
}

func (infoPlist *InfoPlist) BuildNumberString() string {
	return infoPlist.object[buildNumberKey].(string)
}

func (infoPlist *InfoPlist) NextMajor() (string, error) {
	versionString := infoPlist.VersionString()
	version, err := semver.Make(versionString)
	if err != nil {
		return versionString, err
	}

	version.Major += 1
	version.Minor = 0
	version.Patch = 0

	return version.String(), nil
}

func (infoPlist *InfoPlist) IncrementMajor() error {
	versionString, err := infoPlist.NextMajor()
	if err != nil {
		return err
	}

	infoPlist.object[versionKey] = versionString

	return nil
}

func (infoPlist *InfoPlist) NextMinor() (string, error) {
	versionString := infoPlist.VersionString()
	version, err := semver.Make(versionString)
	if err != nil {
		return versionString, err
	}

	version.Minor += 1
	version.Patch = 0

	return version.String(), nil
}

func (infoPlist *InfoPlist) IncrementMinor() error {
	versionString, err := infoPlist.NextMinor()
	if err != nil {
		return err
	}

	infoPlist.object[versionKey] = versionString

	return nil
}

func (infoPlist *InfoPlist) NextPatch() (string, error) {
	versionString := infoPlist.VersionString()
	version, err := semver.Make(versionString)
	if err != nil {
		return versionString, err
	}

	version.Patch += 1

	return version.String(), nil
}

func (infoPlist *InfoPlist) IncrementPatch() error {
	versionString, err := infoPlist.NextPatch()
	if err != nil {
		return err
	}

	infoPlist.object[versionKey] = versionString

	return nil
}

func (infoPlist *InfoPlist) NextBuildNumber() (string, error) {
	buildNumberString := infoPlist.BuildNumberString()
	buildNumber, err := strconv.Atoi(buildNumberString)
	if err != nil {
		return buildNumberString, err
	}

	return strconv.Itoa(buildNumber + 1), nil
}

func (infoPlist *InfoPlist) IncrementBuildNumber() error {
	buildNumberString, err := infoPlist.NextBuildNumber()
	if err != nil {
		return err
	}

	infoPlist.object[buildNumberKey] = buildNumberString

	return nil
}

func (infoPlist *InfoPlist) SetVersion(version string, build string) {
	infoPlist.object[versionKey] = version
	infoPlist.object[buildNumberKey] = build
}

func (infoPlist *InfoPlist) WriteToFile(path string) error {
	bytes, err := plist.MarshalIndent(&infoPlist.object, plist.XMLFormat, "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
