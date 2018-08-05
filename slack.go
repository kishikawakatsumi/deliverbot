package main

import (
	"fmt"
	"strings"

	"github.com/nlopes/slack"
)

const (
	actionBranch      = "branch"
	actionVersion     = "version"
	actionBuildNumber = "buildNumber"
	actionRun         = "run"
	actionCancel      = "cancel"

	callbackID  = "deliver:parameters"
	helpMessage = "```\nUsage:\n\t@applebot deliver\n\t@applebot ping\n\t@applebot help```"
)

type SlackListener struct {
	client         *slack.Client
	botID          string
	channelID      string
	debugChannelID string
}

func (s *SlackListener) ListenAndResponse() {
	rtm := s.client.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(ev); err != nil {
				sugar.Errorf("Failed to handle message: %s", err)
			}
		}
	}
}

func (s *SlackListener) handleMessageEvent(ev *slack.MessageEvent) error {
	if ev.Channel != s.channelID && ev.Channel != s.debugChannelID {
		return nil
	}

	fields := strings.Fields(ev.Msg.Text)
	if len(fields) == 0 || len(fields) > 2 {
		return nil
	}

	mentionToBot := fields[0] == fmt.Sprintf("<@%s>", s.botID)
	if len(fields) == 1 && mentionToBot {
		err := s.respond(ev.Channel, helpMessage)
		return err
	}
	if len(fields) == 2 && mentionToBot && fields[1] == "ping" {
		err := s.respond(ev.Channel, "pong")
		return err
	}
	if len(fields) == 2 && mentionToBot && fields[1] == "help" {
		err := s.respond(ev.Channel, helpMessage)
		return err
	}
	if len(fields) == 2 && mentionToBot && fields[1] == "deliver" {
		buildParameters := BuildParameters{}

		actions, err := branchOptions(buildParameters)
		if err != nil {
			return err
		}
		messageParameters := slack.PostMessageParameters{
			Attachments: []slack.Attachment{
				{
					Text:       "Branch:",
					CallbackID: callbackID,
					Actions:    actions,
				},
			},
		}

		if _, _, err := s.client.PostMessage(ev.Channel, "", messageParameters); err != nil {
			return fmt.Errorf("failed to post message: %s", err)
		}
		return nil
	}

	return nil
}

func branchOptions(parameters BuildParameters) ([]slack.AttachmentAction, error) {
	defaultBranch, err := service.DefaultBranch()
	if err != nil {
		return []slack.AttachmentAction{}, err
	}

	var options []slack.AttachmentActionOption
	branches, err := service.Branches()
	if err != nil {
		return []slack.AttachmentAction{}, err
	}
	for _, branch := range branches {
		if branch.GetName() == *defaultBranch {
			continue
		}
		featureBranchParameters := parameters
		featureBranchParameters.Branch = branch.GetName()
		options = append(options, slack.AttachmentActionOption{
			Text:  branch.GetName(),
			Value: featureBranchParameters.string(),
		})
	}

	defaultBranchParameters := parameters
	defaultBranchParameters.Branch = *defaultBranch
	actions := []slack.AttachmentAction{
		{
			Name:  actionBranch,
			Text:  defaultBranchParameters.Branch,
			Value: defaultBranchParameters.string(),
			Type:  "button",
			Style: "primary",
		},
		{
			Name:    actionBranch,
			Text:    "Other branch...",
			Type:    "select",
			Options: options,
		},
		cancelAction,
	}
	return actions, nil
}

func (s *SlackListener) respond(channel string, text string) error {
	_, _, err := s.client.PostMessage(channel, text, slack.NewPostMessageParameters())
	return fmt.Errorf("failed to post message: %s", err)
}
