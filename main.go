package main

import (
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
	"go.uber.org/zap"
)

type envConfig struct {
	Port                  string `envconfig:"PORT" default:"3000"`
	BotToken              string `envconfig:"BOT_TOKEN" required:"true"`
	VerificationToken     string `envconfig:"VERIFICATION_TOKEN" required:"true"`
	BotID                 string `envconfig:"BOT_ID" required:"true"`
	ChannelID             string `envconfig:"CHANNEL_ID" required:"true"`
	GitHubUsername        string `envconfig:"GITHUB_USERNAME" required:"true"`
	GitHubToken           string `envconfig:"GITHUB_TOKEN" required:"true"`
	GitHubRepositoryOwner string `envconfig:"GITHUB_REPOSITORY_OWNER" required:"true"`
	GitHubRepositoryName  string `envconfig:"GITHUB_REPOSITORY_NAME" required:"true"`
	GitCommitAuthorName   string `envconfig:"GIT_COMMIT_AUTHOR_NAME" required:"true"`
	GitCommitAuthorEmail  string `envconfig:"GIT_COMMIT_AUTHOR_EMAIL" required:"true"`
	InfoPlistPath         string `envconfig:"INFOPLIST_PATH" required:"true"`
}

var (
	logger  *zap.Logger
	sugar   *zap.SugaredLogger
	service *GitHubService
	infoPlistPath string
)

func main() {
	os.Exit(_main(os.Args[1:]))
}

func _main(args []string) int {
	logger, _ = zap.NewDevelopment()
	defer logger.Sync()
	sugar = logger.Sugar()

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		sugar.Errorf("Failed to process env var: %s", err)
		return 1
	}

	repo := GitHubRepository{Owner: env.GitHubRepositoryOwner, Name: env.GitHubRepositoryName}
	author := CommitAuthor{Name: env.GitCommitAuthorName, Email: env.GitCommitAuthorEmail}
	service = NewGitHubService(env.GitHubToken, repo, author)
	infoPlistPath = env.InfoPlistPath

	sugar.Infof("Start slack event listening")
	client := slack.New(env.BotToken)
	slackListener := &SlackListener{
		client:    client,
		botID:     env.BotID,
		channelID: env.ChannelID,
	}
	go slackListener.ListenAndResponse()

	http.Handle("/interaction", interactionHandler{
		verificationToken: env.VerificationToken,
	})

	sugar.Infof("Server listening on :%s", env.Port)
	if err := http.ListenAndServe(":"+env.Port, nil); err != nil {
		sugar.Errorf("%s", err)
		return 1
	}

	return 0
}
