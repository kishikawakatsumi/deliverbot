package main

import (
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type interactionHandler struct {
	slackClient       *slack.Client
	verificationToken string
}

func (h interactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sugar.Errorf("Invalid method: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sugar.Errorf("Failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		sugar.Errorf("Failed to un-escape request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message slack.AttachmentActionCallback
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		sugar.Errorf("Failed to decode json message from slack: %s", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if message.Token != h.verificationToken {
		sugar.Errorf("Invalid token: %s", message.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	action := message.Actions[0]
	parameters := NewBuildParameters(action.Value)
	if action.Value == "" {
		parameters = NewBuildParameters(action.SelectedOptions[0].Value)
	}
	switch action.Name {
	case actionBranch:
		file, err := service.File(parameters.Branch, service.InfoPlistPath)
		if err != nil {
			responseError(w, message.OriginalMessage, "Error occurred.", fmt.Sprintf("%s", err))
			return
		}

		infoPlist, err := NewInfoPlist(file)
		if err != nil {
			responseError(w, message.OriginalMessage, "Error occurred.", fmt.Sprintf("%s", err))
			return
		}
		tempFile, err := ioutil.TempFile("", "applebot-")

		bytes, err := infoPlist.serialized()
		if err != nil {
			responseError(w, message.OriginalMessage, "Error occurred.", fmt.Sprintf("%s", err))
			return
		}
		tempFile.Write(bytes)

		currentVersion := infoPlist.VersionString()
		currentBuildNumber := infoPlist.BuildNumberString()

		nextPatch, err := infoPlist.NextPatch()
		if err != nil {
			responseError(w, message.OriginalMessage, "Error occurred.", fmt.Sprintf("%s", err))
			return
		}
		nextMinor, err := infoPlist.NextMinor()
		if err != nil {
			responseError(w, message.OriginalMessage, "Error occurred.", fmt.Sprintf("%s", err))
			return
		}
		nextMajor, err := infoPlist.NextMajor()
		if err != nil {
			responseError(w, message.OriginalMessage, "Error occurred.", fmt.Sprintf("%s", err))
			return
		}
		nextBuildNumber, err := infoPlist.NextBuildNumber()
		if err != nil {
			responseError(w, message.OriginalMessage, "Error occurred.", fmt.Sprintf("%s", err))
			return
		}

		buildParameters := BuildParameters{
			Branch:             parameters.Branch,
			Version:            "",
			BuildNumber:        "",
			CurrentVersion:     currentVersion,
			CurrentBuildNumber: currentBuildNumber,
			NextPatch:          nextPatch,
			NextMinor:          nextMinor,
			NextMajor:          nextMajor,
			NextBuildNumber:    nextBuildNumber,
			InfoPlist:          tempFile.Name(),
		}

		responseAction(w, message.OriginalMessage, fmt.Sprintf("Branch: `%s`\nVersion:", parameters.Branch), versionOptions(buildParameters))
	case actionVersion:
		responseAction(w, message.OriginalMessage, fmt.Sprintf("Branch: `%s`\nVersion: `%s`\nBuild:", parameters.Branch, parameters.Version), buildNumberOptions(parameters))
	case actionBuildNumber:
		currentVersion := fmt.Sprintf("%s (%s)", parameters.CurrentVersion, parameters.CurrentBuildNumber)
		nextVersion := fmt.Sprintf("%s (%s)", parameters.Version, parameters.BuildNumber)
		responseAction(w, message.OriginalMessage, fmt.Sprintf("Branch: `%s`\nCurrent Version: `%s`\nNext Version: `%s`", parameters.Branch, currentVersion, nextVersion), okCancelOptions(parameters))
	case actionRun:
		bytes, err := ioutil.ReadFile(parameters.InfoPlist)
		if err != nil {
			responseError(w, message.OriginalMessage, "Error occurred.", fmt.Sprintf("%s", err))
			return
		}
		infoPlist, err := NewInfoPlist(bytes)
		if err != nil {
			responseError(w, message.OriginalMessage, "Error occurred.", fmt.Sprintf("%s", err))
			return
		}

		responseMessage(w, message.OriginalMessage, "Running ...", "")

		go func() {
			infoPlist.SetVersion(parameters.Version, parameters.BuildNumber)

			bytes, _ := infoPlist.serialized()

			timestamp := strconv.FormatInt(time.Now().Unix(), 10)
			commitBranch := fmt.Sprintf("release/%s-%s-%s", parameters.Version, parameters.BuildNumber, timestamp)
			title := fmt.Sprintf("Release %s (%s)", parameters.Version, parameters.BuildNumber)
			commitMessage := title

			u, err := service.PushPullRequest(PullRequest{
				TargetBranch:  parameters.Branch,
				CommitBranch:  commitBranch,
				FileContent:   bytes,
				FilePath:      service.InfoPlistPath,
				Title:         title,
				CommitMessage: commitMessage,
			})
			if err != nil {
				sugar.Errorf("Failed to create pull request %s", err)
			} else {
				m := fmt.Sprintf("Releasing `%s (%s)`", infoPlist.VersionString(), infoPlist.BuildNumberString())
				sugar.Infof(m)
				h.slackClient.PostMessage(message.OriginalMessage.Channel, fmt.Sprintf("%s\n%s", m, u), slack.PostMessageParameters{})
			}
		}()
	case actionCancel:
		responseMessage(w, message.OriginalMessage, fmt.Sprintf("Operation canceled by '%s'.", message.User.Name), "")
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func responseMessage(w http.ResponseWriter, original slack.Message, title, value string) {
	original.Attachments[0].Actions = []slack.AttachmentAction{}
	original.Attachments[0].Fields = []slack.AttachmentField{
		{
			Title: title,
			Value: value,
			Short: false,
		},
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&original)
}

func responseAction(w http.ResponseWriter, original slack.Message, text string, actions []slack.AttachmentAction) {
	original.Attachments[0].Text = text
	original.Attachments[0].Actions = actions

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&original)
}

func responseError(w http.ResponseWriter, original slack.Message, title, value string) {
	original.Attachments[0].Actions = []slack.AttachmentAction{}
	original.Attachments[0].Fields = []slack.AttachmentField{
		{
			Title: title,
			Value: value,
			Short: false,
		},
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(&original)
}

func versionOptions(parameters BuildParameters) []slack.AttachmentAction {
	parameters.Version = parameters.CurrentVersion
	currentVersionAction := slack.AttachmentAction{
		Name:  actionVersion,
		Text:  parameters.Version,
		Value: parameters.string(),
		Type:  "button",
		Style: "primary",
	}

	parameters.Version = parameters.NextPatch
	patchVersionAction := slack.AttachmentAction{
		Name:  actionVersion,
		Text:  parameters.Version,
		Value: parameters.string(),
		Type:  "button",
	}

	parameters.Version = parameters.NextMinor
	minorVersionAction := slack.AttachmentAction{
		Name:  actionVersion,
		Text:  parameters.Version,
		Value: parameters.string(),
		Type:  "button",
	}

	parameters.Version = parameters.NextMajor
	majorVersionAction := slack.AttachmentAction{
		Name:  actionVersion,
		Text:  parameters.Version,
		Value: parameters.string(),
		Type:  "button",
	}
	actions := []slack.AttachmentAction{
		currentVersionAction,
		patchVersionAction,
		minorVersionAction,
		majorVersionAction,
		cancelAction(),
	}
	return actions
}

func buildNumberOptions(parameters BuildParameters) []slack.AttachmentAction {
	parameters.BuildNumber = parameters.NextBuildNumber
	currentBuildNumberAction := slack.AttachmentAction{
		Name:  actionBuildNumber,
		Text:  parameters.BuildNumber,
		Value: parameters.string(),
		Type:  "button",
		Style: "primary",
	}

	nextBuildNumber := parameters.NextBuildNumber
	number, _ := strconv.Atoi(nextBuildNumber)
	number++
	var options []slack.AttachmentActionOption
	for i := number; i <= number+5; i++ {
		buildNumber := strconv.Itoa(i)
		parameters.BuildNumber = buildNumber
		options = append(options, slack.AttachmentActionOption{
			Text:  buildNumber,
			Value: parameters.string(),
		})
	}
	actions := []slack.AttachmentAction{
		currentBuildNumberAction,
		{
			Name:    actionBuildNumber,
			Text:    "Build number",
			Type:    "select",
			Options: options,
		},
		cancelAction(),
	}
	return actions
}

func okCancelOptions(parameters BuildParameters) []slack.AttachmentAction {
	okAction := slack.AttachmentAction{
		Name:  actionRun,
		Text:  "　OK　",
		Value: parameters.string(),
		Type:  "button",
		Style: "primary",
	}
	actions := []slack.AttachmentAction{
		okAction,
		cancelAction(),
	}
	return actions
}

func cancelAction() slack.AttachmentAction {
	return slack.AttachmentAction{
		Name:  actionCancel,
		Text:  "Cancel",
		Value: "cancel",
		Type:  "button",
		Style: "danger",
	}
}
