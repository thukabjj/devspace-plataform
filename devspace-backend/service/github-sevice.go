package service

import (
	"bytes"
	"context"
	"devspace-backend/dto"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
)

func CreateGithubRepository(productRepositoryDTO dto.ProjectRepositoryDTO) error {
	// Check if required environment variables are set
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}

	// Create the GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Create the new repository
	repo := &github.Repository{
		Name:        github.String(productRepositoryDTO.Name),
		Description: github.String(productRepositoryDTO.Description),
		Language:    github.String(productRepositoryDTO.Language),
		Private:     github.Bool(true),
	}
	createdRepo, _, err := client.Repositories.Create(ctx, "", repo)
	if err != nil {
		return fmt.Errorf("failed to create repository: %v", err)
	}

	// Clone the template project from GitHub
	templateURL := "https://github.com/thukabjj/devspace-plataform.git"
	templatePath := filepath.Join(os.TempDir(), "devspace-plataform")
	cloneOpt := &git.CloneOptions{
		URL:      templateURL,
		Progress: os.Stdout,
	}

	// Delete the existing repository
	err = os.RemoveAll(templatePath)
	if err != nil {
		return fmt.Errorf("failed to remove existing repository: %v", err)
	}

	// Clone the repository
	r, err := git.PlainCloneContext(ctx, templatePath, false, cloneOpt)
	if err != nil {
		if err != git.ErrRepositoryAlreadyExists {
			return fmt.Errorf("failed to clone template repository: %v", err)
		}
		r, err = git.PlainOpen(templatePath)
		if err != nil {
			return fmt.Errorf("failed to open cloned repository: %v", err)
		}
	}

	// Add, commit, and push the modified files to the new repository
	worktree, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}

	// Replace the <module_name> placeholder with the actual module name
	moduleName := productRepositoryDTO.Name
	err = filepath.Walk(filepath.Join(templatePath, "project-template", "golang"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		content = bytes.ReplaceAll(content, []byte("<module_name>"), []byte(moduleName))
		relPath, err := filepath.Rel(templatePath, path)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(path, content, 0666)
		if err != nil {
			return err
		}
		_, err = worktree.Add(relPath)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to replace <module_name> placeholder: %v", err)
	}

	// Add and commit the changes
	_, err = worktree.Add(".")
	if err != nil {
		return fmt.Errorf("failed to add files to index: %v", err)
	}

	// Create the commit
	commitMsg := fmt.Sprintf("Initial commit for %s", moduleName)
	_, err = worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Your Name",
			Email: "your.email@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit files: %v", err)
	}

	// Push the changes to the new repository
	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to push files to repository: %v", err)
	}

	fmt.Printf("Created repository: %s\n", createdRepo.GetName())
	fmt.Println("Successfully created new repository and pushed template to it!")

	return nil
}
