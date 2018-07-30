package main

import (
	"github.com/nlopes/slack"
	"go.uber.org/zap"
	"net/http"
	"os"
)

var (
	logger        *zap.Logger
	sugar         *zap.SugaredLogger
	service       *GitHubService
)

func main() {
	os.Exit(_main(os.Args[1:]))
}

func _main(_ []string) int {
	logger, _ = zap.NewDevelopment()
	defer logger.Sync()
	sugar = logger.Sugar()

	//var env envConfig
	config, err := LoadConfig()
	if err != nil {
		sugar.Errorf("Failed to load system config: %s", err)
		return 1
	}

	repo := GitHubRepository{Owner: config.GitHubRepositoryOwner, Name: config.GitHubRepositoryName}
	author := CommitAuthor{Name: config.GitCommitAuthorName, Email: config.GitCommitAuthorEmail}
	service = NewGitHubService(config.GitHubToken, repo, author)

	sugar.Infof("Start slack event listening")
	client := slack.New(config.BotToken)
	slackListener := &SlackListener{
		client:    client,
		botID:     config.BotID,
		channelID: config.ChannelID,
	}
	go slackListener.ListenAndResponse()

	http.Handle("/interaction", interactionHandler{
		verificationToken: config.VerificationToken,
	})

	sugar.Infof("Server listening on :%s", config.Port)
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		sugar.Errorf("%s", err)
		return 1
	}

	return 0
}
