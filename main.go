package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
)

type envConfig struct {
	Port                 string `envconfig:"PORT" default:"3000"`
	BotToken             string `envconfig:"BOT_TOKEN" required:"true"`
	VerificationToken    string `envconfig:"VERIFICATION_TOKEN" required:"true"`
	BotID                string `envconfig:"BOT_ID" required:"true"`
	ChannelID            string `envconfig:"CHANNEL_ID" required:"true"`
	GitHubUsername       string `envconfig:"GITHUB_USERNAME" required:"false"`
	GitHubToken          string `envconfig:"GITHUB_TOKEN" required:"false"`
	GitHubRepositorySlug string `envconfig:"GITHUB_REPOSITORY_SLUG" required:"true"`
	GitCloneLocalPath    string `envconfig:"GIT_CLONE_LOCAL_PATH" required:"true"`
	GitCommitAuthorName  string `envconfig:"GIT_COMMIT_AUTHOR_NAME" required:"true"`
	GitCommitAuthorEmail string `envconfig:"GIT_COMMIT_AUTHOR_EMAIL" required:"true"`
	InfoPlistPath        string `envconfig:"INFOPLIST_PATH" required:"true"`
	GitBranches          string `envconfig:"GIT_BRANCHES" required:"false"`
}

func main() {
	os.Exit(_main(os.Args[1:]))
}

func _main(args []string) int {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		return 1
	}

	repositoryOptions := RepositoryOptions{
		path: env.GitCloneLocalPath,
		credential: Credential{
			Username: env.GitHubUsername,
			Token:    env.GitHubToken,
		},
		slug: env.GitHubRepositorySlug,
		author: Author{
			Name:  env.GitCommitAuthorName,
			Email: env.GitCommitAuthorEmail,
		},
		infoPlistPath: env.InfoPlistPath,
		branches:      env.GitBranches,
	}

	log.Printf("[INFO] Start slack event listening")
	client := slack.New(env.BotToken)
	slackListener := &SlackListener{
		client:            client,
		botID:             env.BotID,
		channelID:         env.ChannelID,
		repositoryOptions: repositoryOptions,
	}
	go slackListener.ListenAndResponse()

	http.Handle("/interaction", interactionHandler{
		verificationToken: env.VerificationToken,
		channelID:         env.ChannelID,
		repositoryOptions: repositoryOptions,
	})

	log.Printf("[INFO] Server listening on :%s", env.Port)
	if err := http.ListenAndServe(":"+env.Port, nil); err != nil {
		log.Printf("[ERROR] %s", err)
		return 1
	}

	return 0
}
