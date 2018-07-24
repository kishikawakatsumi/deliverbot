package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"strings"
	"encoding/json"
)

const (
	releaseMasterBranch = "releaseMasterBranch"
	releaseOtherBranch  = "releaseOtherBranch"
	currentVersion      = "currentVersion"
	incrementPatch      = "incrementPatch"
	incrementMinor      = "incrementMinor"
	incrementMajor      = "incrementMajor"
	defaultBuildNumber  = "defaultBuildNumber"
	customBuildNumber   = "customBuildNumber"
)

type SlackListener struct {
	client            *slack.Client
	botID             string
	channelID         string
	repositoryOptions RepositoryOptions
}

type RepositoryOptions struct {
	path          string
	credential    Credential
	slug          string
	author        Author
	infoPlistPath string
	branches      string
}

func (s *SlackListener) ListenAndResponse() {
	rtm := s.client.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(ev); err != nil {
				log.Printf("[ERROR] Failed to handle message: %s", err)
			}
		}
	}
}

func (s *SlackListener) handleMessageEvent(ev *slack.MessageEvent) error {
	if ev.Channel != s.channelID {
		return nil
	}

	helpMessage := "```\nCommand:\n\t@applebot deliver\n```"

	fields := strings.Fields(ev.Msg.Text)
	if len(fields) == 0 {
		return nil
	}
	if len(fields) == 1 && fields[0] == fmt.Sprintf("<@%s>", s.botID) {
		// Show help
		if _, _, err := s.client.PostMessage(ev.Channel, helpMessage, slack.NewPostMessageParameters()); err != nil {
			return fmt.Errorf("failed to post message: %s", err)
		}
		return nil
	}
	if len(fields) == 2 && fields[0] == fmt.Sprintf("<@%s>", s.botID) {
		if fields[1] == "deliver" {
			var options []slack.AttachmentActionOption
			var branches []string
			if err := json.Unmarshal([]byte(s.repositoryOptions.branches), &branches); err != nil {
				return err
			}
			for _, branch := range branches {
				options = append(options, slack.AttachmentActionOption{
					Text:  branch,
					Value: branch,
				})
			}
			attachment := slack.Attachment{
				Text:       "Which branch?",
				CallbackID: "select:branch",
				Actions: []slack.AttachmentAction{
					{
						Name:  releaseMasterBranch,
						Text:  "master",
						Value: "master",
						Type:  "button",
						Style: "primary",
					},
					{
						Name:    releaseOtherBranch,
						Type:    "select",
						Options: options,
					},
				},
			}
			parameters := slack.PostMessageParameters{
				Attachments: []slack.Attachment{
					attachment,
				},
			}

			if _, _, err := s.client.PostMessage(ev.Channel, "", parameters); err != nil {
				return fmt.Errorf("failed to post message: %s", err)
			}
			return nil
		}
		if fields[1] == "ping" {
			if _, _, err := s.client.PostMessage(ev.Channel, "pong", slack.NewPostMessageParameters()); err != nil {
				return fmt.Errorf("failed to post message: %s", err)
			}
			return nil
		}
		if fields[1] == "help" {
			if _, _, err := s.client.PostMessage(ev.Channel, helpMessage, slack.NewPostMessageParameters()); err != nil {
				return fmt.Errorf("failed to post message: %s", err)
			}
			return nil
		}
	}

	return nil
}
