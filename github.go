package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type GitHubService struct {
	Repository    GitHubRepository
	Author        CommitAuthor
	InfoPlistPath string
	Client        *github.Client
}

type GitHubRepository struct {
	Owner string
	Name  string
}

type CommitAuthor struct {
	Name  string
	Email string
}

type PullRequest struct {
	TargetBranch  string
	CommitBranch  string
	FileContent   []byte
	FilePath      string
	Title         string
	CommitMessage string
}

func NewGitHubService(token string, repo GitHubRepository, author CommitAuthor, infoPlistPath string) *GitHubService {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &GitHubService{
		Repository:    repo,
		Author:        author,
		Client:        client,
		InfoPlistPath: infoPlistPath,
	}
}

func (g *GitHubService) DefaultBranch() (*string, error) {
	repo, _, err := g.Client.Repositories.Get(context.Background(), g.Repository.Owner, g.Repository.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub repository: %s", err)
	}
	defaultBranch := repo.GetDefaultBranch()
	return &defaultBranch, nil
}

func (g *GitHubService) Branches() ([]github.Branch, error) {
	branches, _, err := g.Client.Repositories.ListBranches(context.Background(), g.Repository.Owner, g.Repository.Name, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub branches: %s", err)
	}
	return filter(branches, func(branch github.Branch) bool { return !strings.Contains(branch.GetName(), "/") }), nil
}

func (g *GitHubService) File(branch, path string) ([]byte, error) {
	file, err := g.Client.Repositories.DownloadContents(context.Background(), g.Repository.Owner, g.Repository.Name, path, &github.RepositoryContentGetOptions{
		Ref: branch,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from GitHub: %s", err)
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from GitHub: %s", err)
	}

	return bytes, nil
}

func filter(vs []*github.Branch, f func(github.Branch) bool) []github.Branch {
	vsf := make([]github.Branch, 0)
	for _, v := range vs {
		if f(*v) {
			vsf = append(vsf, *v)
		}
	}
	return vsf
}

// CreateBranch returns the commit branch reference object if it exists or creates it
// from the base branch before returning it.
func (g *GitHubService) CreateBranch(from, to string) (ref *github.Reference, err error) {
	if ref, _, err = g.Client.Git.GetRef(context.Background(), g.Repository.Owner, g.Repository.Name, fmt.Sprintf("refs/heads/%s", to)); err == nil {
		return ref, nil
	}

	var baseRef *github.Reference
	if baseRef, _, err = g.Client.Git.GetRef(context.Background(), g.Repository.Owner, g.Repository.Name, fmt.Sprintf("refs/heads/%s", from)); err != nil {
		return nil, err
	}
	newRef := &github.Reference{Ref: github.String(fmt.Sprintf("refs/heads/%s", to)), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = g.Client.Git.CreateRef(context.Background(), g.Repository.Owner, g.Repository.Name, newRef)
	return ref, err
}

// GetTree generates the tree to commit based on the given files and the commit
// of the ref you got in getRef.
func (g *GitHubService) CreateTree(ref *github.Reference, content []byte, path string) (tree *github.Tree, err error) {
	entries := []github.TreeEntry{}
	entries = append(entries, github.TreeEntry{Path: github.String(path), Type: github.String("blob"), Content: github.String(string(content)), Mode: github.String("100644")})
	tree, _, err = g.Client.Git.CreateTree(context.Background(), g.Repository.Owner, g.Repository.Name, *ref.Object.SHA, entries)
	return tree, err
}

// PushCommit creates the commit in the given reference using the given tree.
func (g *GitHubService) PushCommit(ref *github.Reference, tree *github.Tree, commitMessage string) (err error) {
	// Get the parent commit to attach the commit to.
	parent, _, err := g.Client.Repositories.GetCommit(context.Background(), g.Repository.Owner, g.Repository.Name, *ref.Object.SHA)
	if err != nil {
		return err
	}
	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA

	// Create the commit using the tree.
	date := time.Now()
	author := &github.CommitAuthor{Date: &date, Name: &g.Author.Name, Email: &g.Author.Email}
	commit := &github.Commit{Author: author, Message: &commitMessage, Tree: tree, Parents: []github.Commit{*parent.Commit}}
	newCommit, _, err := g.Client.Git.CreateCommit(context.Background(), g.Repository.Owner, g.Repository.Name, commit)
	if err != nil {
		return err
	}

	// Attach the commit to the master branch.
	ref.Object.SHA = newCommit.SHA
	_, _, err = g.Client.Git.UpdateRef(context.Background(), g.Repository.Owner, g.Repository.Name, ref, false)
	return err
}

// CreatePR creates a pull request. Based on: https://godoc.org/github.com/google/go-github/github#example-PullRequestsService-Create
func (g *GitHubService) CreatePullRequest(targetBranch, commitBranch, title, description string) (*github.PullRequest, error) {
	branch := fmt.Sprintf("%s:%s", g.Repository.Owner, commitBranch)

	newPR := &github.NewPullRequest{
		Title:               &title,
		Head:                &branch,
		Base:                &targetBranch,
		Body:                &description,
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := g.Client.PullRequests.Create(context.Background(), g.Repository.Owner, g.Repository.Name, newPR)
	if err != nil {
		return pr, err
	}

	sugar.Infof("PR created: %s\n", pr.GetHTMLURL())
	return pr, nil
}

func (g *GitHubService) PushPullRequest(pullRequest PullRequest) (*string, error) {
	ref, err := g.CreateBranch(pullRequest.TargetBranch, pullRequest.CommitBranch)
	if err != nil {
		sugar.Errorf("Unable to get/create the commit reference: %s\n", err)
		return nil, err
	}
	if ref == nil {
		sugar.Errorf("No error where returned but the reference is nil")
		return nil, err
	}

	tree, err := g.CreateTree(ref, pullRequest.FileContent, pullRequest.FilePath)
	if err != nil {
		sugar.Errorf("Unable to create the tree based on the provided files: %s\n", err)
		return nil, err
	}

	if err := g.PushCommit(ref, tree, pullRequest.CommitMessage); err != nil {
		sugar.Errorf("Unable to create the commit: %s\n", err)
		return nil, err
	}

	pr, err := g.CreatePullRequest(pullRequest.TargetBranch, pullRequest.CommitBranch, pullRequest.Title, pullRequest.CommitMessage)
	if err != nil {
		log.Fatalf("Error while creating the pull request: %s", err)
		return nil, err
	}

	u := pr.GetHTMLURL()
	return &u, nil
}

func (g *GitHubService) LatestTag() (*github.RepositoryTag, error) {
	tags, _, err := g.Client.Repositories.ListTags(context.Background(), g.Repository.Owner, g.Repository.Name, &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub tags: %s", err)
	}
	if len(tags) == 0 {
		return nil, fmt.Errorf("no tags")
	}
	return tags[0], nil
}

func (g *GitHubService) Commits(base string, head string) ([]github.RepositoryCommit, error) {
	commitsComparison, _, err := g.Client.Repositories.CompareCommits(context.Background(), g.Repository.Owner, g.Repository.Name, base, head)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub commits: %s", err)
	}
	return commitsComparison.Commits, nil
}
