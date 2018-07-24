package main

import (
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type interactionHandler struct {
	slackClient       *slack.Client
	verificationToken string
	channelID         string
	repositoryOptions RepositoryOptions
}

type callback func(error)

type AppVersion struct {
	Version string
	Build   string
}

func (h interactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] Invalid method: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		log.Printf("[ERROR] Failed to unespace request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message slack.AttachmentActionCallback
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		log.Printf("[ERROR] Failed to decode json message from slack: %s", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Only accept message from slack with valid token
	if message.Token != h.verificationToken {
		log.Printf("[ERROR] Invalid token: %s", message.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	action := message.Actions[0]
	switch action.Name {
	case releaseMasterBranch:
		err := h.checkout("master")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = h.showVersionOptions(w, message.OriginalMessage)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case releaseOtherBranch:
		err := h.checkout(action.SelectedOptions[0].Value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = h.showVersionOptions(w, message.OriginalMessage)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case currentVersion:
		fallthrough
	case incrementPatch:
		fallthrough
	case incrementMinor:
		fallthrough
	case incrementMajor:
		err := h.showBuildNumberOptions(w, message.OriginalMessage, action)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case defaultBuildNumber:
		var appVersion = AppVersion{}
		if err := json.Unmarshal([]byte(action.Value), &appVersion); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		versionString := fmt.Sprintf("%s (%s)", appVersion.Version, appVersion.Build)
		go h.commitAndPush(appVersion, func(err error) {
			if err != nil {
				return
			}
		})

		title := fmt.Sprintf("Branch: `master`, Version: `%s`", versionString)
		responseMessage(w, message.OriginalMessage, title, "")
	case customBuildNumber:
		var appVersion = AppVersion{}
		if err := json.Unmarshal([]byte(action.SelectedOptions[0].Value), &appVersion); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		versionString := fmt.Sprintf("%s (%s)", appVersion.Version, appVersion.Build)
		go h.commitAndPush(appVersion, func(err error) {
			if err != nil {
				return
			}
		})

		title := fmt.Sprintf("Delivering `%s` ...", versionString)
		responseMessage(w, message.OriginalMessage, title, "")
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h interactionHandler) checkout(branch string) error {
	root := h.repositoryOptions.path

	repository, err := NewRepository(root, h.repositoryOptions.slug, &h.repositoryOptions.credential)
	if err != nil {
		return err
	}

	err = repository.Checkout(fmt.Sprintf("refs/heads/%s", branch), false)
	if err != nil {
		return err
	}

	return nil
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

func (h interactionHandler) showVersionOptions(w http.ResponseWriter, original slack.Message) error {
	root := h.repositoryOptions.path
	credential := h.repositoryOptions.credential

	_, err := NewRepository(root, h.repositoryOptions.slug, &credential)
	if err != nil {
		return err
	}

	infoPlist, err := NewInfoPlist(fmt.Sprintf("%s/%s", root, h.repositoryOptions.infoPlistPath))
	if err != nil {
		return err
	}

	nextMajor, err := infoPlist.NextMajor()
	if err != nil {
		return err
	}
	nextMinor, err := infoPlist.NextMinor()
	if err != nil {
		return err
	}
	nextPatch, err := infoPlist.NextPatch()
	if err != nil {
		return err
	}
	nextBuildNumber, err := infoPlist.NextBuildNumber()
	if err != nil {
		return err
	}

	original.Attachments[0].Text = "Choose next version:"
	original.Attachments[0].Actions = []slack.AttachmentAction{
		{
			Name:  currentVersion,
			Text:  infoPlist.VersionString(),
			Value: fmt.Sprintf(`{"version": "%s", "build": "%s"}`, infoPlist.VersionString(), nextBuildNumber),
			Type:  "button",
			Style: "primary",
		},
		{
			Name:  incrementPatch,
			Text:  nextPatch,
			Value: fmt.Sprintf(`{"version": "%s", "build": "%s"}`, nextPatch, nextBuildNumber),
			Type:  "button",
		},
		{
			Name:  incrementMinor,
			Text:  nextMinor,
			Value: fmt.Sprintf(`{"version": "%s", "build": "%s"}`, nextMinor, nextBuildNumber),
			Type:  "button",
		},
		{
			Name:  incrementMajor,
			Text:  nextMajor,
			Value: fmt.Sprintf(`{"version": "%s", "build": "%s"}`, nextMajor, nextBuildNumber),
			Type:  "button",
		},
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&original)

	return nil
}

func (h interactionHandler) showBuildNumberOptions(w http.ResponseWriter, original slack.Message, action slack.AttachmentAction) error {
	var appVersion = AppVersion{}
	if err := json.Unmarshal([]byte(action.Value), &appVersion); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	buildNumber, err := strconv.Atoi(appVersion.Build)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	var options []slack.AttachmentActionOption
	for i := 0; i < 5; i++ {
		build := strconv.Itoa(buildNumber + i + 1)
		options = append(options, slack.AttachmentActionOption{
			Text:  build,
			Value: fmt.Sprintf(`{"version": "%s", "build": "%s"}`, appVersion.Version, build),
		})
	}

	original.Attachments[0].Text = fmt.Sprintf("Version: `%s`\nBuild number:", appVersion.Version)
	original.Attachments[0].Actions = []slack.AttachmentAction{
		{
			Name:  defaultBuildNumber,
			Text:  appVersion.Build,
			Value: fmt.Sprintf(`{"version": "%s", "build": "%s"}`, appVersion.Version, appVersion.Build),
			Type:  "button",
			Style: "primary",
		},
		{
			Name:    customBuildNumber,
			Type:    "select",
			Options: options,
		},
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&original)

	return nil
}

func (h interactionHandler) commitAndPush(version AppVersion, completion callback) {
	root := h.repositoryOptions.path

	repository, err := NewRepository(root, h.repositoryOptions.slug, &h.repositoryOptions.credential)
	if err != nil {
		completion(err)
		return
	}

	branch := fmt.Sprintf("refs/heads/test/%s-%s", version.Version, version.Build)
	err = repository.Checkout(branch, true)
	if err != nil {
		completion(err)
		return
	}

	infoPlist, err := NewInfoPlist(fmt.Sprintf("%s/%s", root, h.repositoryOptions.infoPlistPath))
	if err != nil {
		completion(err)
		return
	}

	infoPlist.SetVersion(version.Version, version.Build)

	err = infoPlist.WriteToFile(infoPlist.Path)
	if err != nil {
		completion(err)
		return
	}

	err = repository.Add(h.repositoryOptions.infoPlistPath)
	if err != nil {
		completion(err)
		return
	}

	err = repository.Commit(h.repositoryOptions.author)
	if err != nil {
		completion(err)
		return
	}

	err = repository.Push()
	if err != nil {
		completion(err)
		return
	}

	completion(nil)
}
