package main

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"time"
)

type Credential struct {
	Username string
	Token    string
}

type Author struct {
	Name  string
	Email string
}

type Repository struct {
	repository *git.Repository
}

func NewRepository(path string, slug string, credential *Credential) (*Repository, error) {
	repo, err := openRepository(path)
	if err != nil {
		repo, err = cloneRepository(path, slug, credential)
		if err != nil {
			return nil, err
		}
	}

	return &Repository{repository: repo}, nil
}

func openRepository(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}

func cloneRepository(path string, slug string, credential *Credential) (*git.Repository, error) {
	var repositoryURL string
	if credential == nil {
		repositoryURL = fmt.Sprintf("https://github.com/%s.git", slug)
	} else {
		repositoryURL = fmt.Sprintf("https://%s:%s@github.com/%s.git", credential.Username, credential.Token, slug)
	}
	return git.PlainClone(path, false, &git.CloneOptions{
		URL: repositoryURL,
	})
}

func (repository *Repository) Checkout(branch string, create bool) error {
	worktree, err := repository.repository.Worktree()
	if err != nil {
		return err
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(branch),
		Create: create,
	})
	if err != nil {
		return err
	}

	return nil
}

func (repository *Repository) Add(path string) error {
	worktree, err := repository.repository.Worktree()
	if err != nil {
		return err
	}

	_, err = worktree.Add(path)
	if err != nil {
		return err
	}

	return nil
}

func (repository *Repository) Commit(author Author) error {
	worktree, err := repository.repository.Worktree()
	if err != nil {
		return err
	}

	_, err = worktree.Commit("Test", &git.CommitOptions{
		Author: &object.Signature{
			Name:  author.Name,
			Email: author.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (repository *Repository) Push() error {
	err := repository.repository.Push(&git.PushOptions{})
	if err != nil {
		return err
	}

	return nil
}
