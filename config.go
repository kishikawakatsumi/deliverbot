package main

import (
	"github.com/kelseyhightower/envconfig"
	toml "github.com/sioncojp/tomlssm"
)

type Config struct {
	BotToken              string
	VerificationToken     string
	BotID                 string
	ChannelID             string
	DebugChannelID        string
	GitHubUsername        string
	GitHubToken           string
	GitHubRepositoryOwner string
	GitHubRepositoryName  string
	GitCommitAuthorName   string
	GitCommitAuthorEmail  string
	InfoPlistPath         string
}

type envConfig struct {
	BotToken              string `envconfig:"BOT_TOKEN"`
	VerificationToken     string `envconfig:"VERIFICATION_TOKEN"`
	BotID                 string `envconfig:"BOT_ID"`
	ChannelID             string `envconfig:"CHANNEL_ID"`
	DebugChannelID        string `envconfig:"DEBUG_CHANNEL_ID"`
	GitHubUsername        string `envconfig:"GITHUB_USERNAME"`
	GitHubToken           string `envconfig:"GITHUB_TOKEN"`
	GitHubRepositoryOwner string `envconfig:"GITHUB_REPOSITORY_OWNER"`
	GitHubRepositoryName  string `envconfig:"GITHUB_REPOSITORY_NAME"`
	GitCommitAuthorName   string `envconfig:"GIT_COMMIT_AUTHOR_NAME"`
	GitCommitAuthorEmail  string `envconfig:"GIT_COMMIT_AUTHOR_EMAIL"`
	InfoPlistPath         string `envconfig:"INFOPLIST_PATH"`
}

type tomlConfig struct {
	BotToken              string `toml:"bot_token"`
	VerificationToken     string `toml:"verification_token"`
	BotID                 string `toml:"bot_id"`
	ChannelID             string `toml:"channel_id"`
	DebugChannelID        string `toml:"debug_channel_id"`
	GitHubUsername        string `toml:"github_username"`
	GitHubToken           string `toml:"github_token"`
	GitHubRepositoryOwner string `toml:"github_repository_owner"`
	GitHubRepositoryName  string `toml:"github_repository_name"`
	GitCommitAuthorName   string `toml:"github_commit_author_name"`
	GitCommitAuthorEmail  string `toml:"github_commit_author_email"`
	InfoPlistPath         string `toml:"infoplist_path"`
}

func LoadConfig(path, region string) (*Config, error) {
	var config Config

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		sugar.Errorf("Failed to process env var: %s", err)
		return nil, err
	}

	tc, err := loadToml(path, region)
	if err != nil {
		sugar.Errorf("Failed to load 'config.toml': %s", err)
		return nil, err
	}

	config.BotToken = tc.BotToken
	if env.BotToken != "" {
		config.BotToken = env.BotToken
	}
	config.VerificationToken = tc.VerificationToken
	if env.VerificationToken != "" {
		config.VerificationToken = env.VerificationToken
	}
	config.BotID = tc.BotID
	if env.BotID != "" {
		config.BotID = env.BotID
	}
	config.ChannelID = tc.ChannelID
	if env.ChannelID != "" {
		config.ChannelID = env.ChannelID
	}
	config.DebugChannelID = tc.DebugChannelID
	if env.DebugChannelID != "" {
		config.DebugChannelID = env.DebugChannelID
	}
	config.GitHubUsername = tc.GitHubUsername
	if env.GitHubUsername != "" {
		config.GitHubUsername = env.GitHubUsername
	}
	config.GitHubToken = tc.GitHubToken
	if env.GitHubToken != "" {
		config.GitHubToken = env.GitHubToken
	}
	config.GitHubRepositoryOwner = tc.GitHubRepositoryOwner
	if env.GitHubRepositoryOwner != "" {
		config.GitHubRepositoryOwner = env.GitHubRepositoryOwner
	}
	config.GitHubRepositoryName = tc.GitHubRepositoryName
	if env.GitHubRepositoryName != "" {
		config.GitHubRepositoryName = env.GitHubRepositoryName
	}
	config.GitCommitAuthorName = tc.GitCommitAuthorName
	if env.GitCommitAuthorName != "" {
		config.GitCommitAuthorName = env.GitCommitAuthorName
	}
	config.GitCommitAuthorEmail = tc.GitCommitAuthorEmail
	if env.GitCommitAuthorEmail != "" {
		config.GitCommitAuthorEmail = env.GitCommitAuthorEmail
	}
	config.InfoPlistPath = tc.InfoPlistPath
	if env.InfoPlistPath != "" {
		config.InfoPlistPath = env.InfoPlistPath
	}

	return &config, nil
}

func loadToml(path , region string) (*tomlConfig, error) {
	var config tomlConfig
	if _, err := toml.DecodeFile(path, &config, region); err != nil {
		return nil, err
	}
	return &config, nil
}
