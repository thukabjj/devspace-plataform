package service

import (
	"bytes"
	"context"
	"devspace-backend/dto"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
)

/**
* TODO:
* - In the creation of the new github repository is set the branch master and should be renamed to main instead master.
* - Remove the temporary template folder after move all the files to the temporary new repository folder.
* - After push in the new repository the initial commit should be deleted the temporary folder of the new repository.
* - Should covered with unit tests
 */
func CreateGithubRepository(productRepositoryDTO dto.ProjectRepositoryDTO) error {
	// Check if required environment variables are set
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}
	ctx := context.Background()

	// Create the GitHub Client
	client := initGitHubClient(ctx, token)

	// Create the new repository
	createdRepo, err := createNewGitHubRepository(ctx, client, productRepositoryDTO)
	if err != nil {
		return err
	}

	templateURL := "https://github.com/thukabjj/devspace-plataform.git"
	templateURLNew := *createdRepo.CloneURL
	templatePathNew := filepath.Join(os.TempDir(), *createdRepo.Name)
	templatePath := filepath.Join(os.TempDir(), "devspace-plataform")

	// Clone the template project from GitHub
	err = cloneInTempDirTemplateRepository(ctx, templateURL, templatePath)
	if err != nil {
		return err
	}

	// Create temporary directory with name of the new repository
	err = createTempDirTemplateRepository(createdRepo, templatePathNew)
	if err != nil {
		return err
	}

	// Initialize a new repository in the template path
	newRepository, err := initializateGitInTemporaryDirectoryFromNewProject(templatePathNew)
	if err != nil {
		return err
	}

	// Add the remote URL
	err = setRemoteGitURLInTemporaryDirectoryFromNewProject(newRepository, templateURLNew)
	if err != nil {
		return err
	}
	// Move files from Temporary template directory to temporary new github repository directory
	errMov := moveTemplateToTempFolder(templatePath, templatePathNew)

	if errMov != nil {
		return errMov
	}

	// Add, commit, and push the modified files to the new repository
	worktreeNew, errNew := newRepository.Worktree()
	if errNew != nil {
		return fmt.Errorf("failed to get worktree of the new repository: %v", err)
	}

	// Replace the <module_name> placeholder with the actual module name
	moduleName := productRepositoryDTO.Name
	err = replaceDefaultNameStructureOfTheTemplateToNewRepository(templatePathNew, moduleName, worktreeNew)
	if err != nil {
		return err
	}

	// Add and commit the changes
	err = addChangesInTheGithubOfNewRepository(worktreeNew)
	if err != nil {
		return err
	}

	// Create the commit
	err = commitChangesInTheGithubOfNewRepository(moduleName, worktreeNew)
	if err != nil {
		return err
	}

	err = pushChangesInTheGithubOfNewRepository(token, newRepository)
	if err != nil {
		return err
	}

	fmt.Printf("Created repository: %s\n", createdRepo.GetName())
	fmt.Println("Successfully created new repository and pushed template to it!")

	return nil
}

func initGitHubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func createNewGitHubRepository(ctx context.Context, client *github.Client, productRepositoryDTO dto.ProjectRepositoryDTO) (*github.Repository, error) {
	repo := &github.Repository{
		Name:        github.String(productRepositoryDTO.Name),
		Description: github.String(productRepositoryDTO.Description),
		Language:    github.String(productRepositoryDTO.Language),
		Private:     github.Bool(true),
	}
	createdRepo, _, err := client.Repositories.Create(ctx, "", repo)

	return createdRepo, err
}

func cloneInTempDirTemplateRepository(ctx context.Context, templateURL string, templatePath string) error {

	cloneOpt := &git.CloneOptions{
		URL:      templateURL,
		Progress: os.Stdout,
	}

	// Clone the repository template project from GitHub
	_, errOr := git.PlainCloneContext(ctx, templatePath, false, cloneOpt)
	if errOr != nil {
		if errOr != git.ErrRepositoryAlreadyExists {
			return fmt.Errorf("failed to clone template repository: %v", errOr)
		}
		_, errOr = git.PlainOpen(templatePath)
		if errOr != nil {
			return fmt.Errorf("failed to open cloned repository: %v", errOr)
		}
	}
	return nil
}

func createTempDirTemplateRepository(createdRepo *github.Repository, templatePathNew string) error {
	err := os.MkdirAll(templatePathNew, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create temp folder for new project: %v", err)
	}
	return nil
}

func initializateGitInTemporaryDirectoryFromNewProject(templatePathNew string) (*git.Repository, error) {
	newRepository, err := git.PlainInit(templatePathNew, false)
	if err != nil {
		return nil, fmt.Errorf("failed to init git in the temporary folder from new repository: %v", err)
	}
	return newRepository, nil
}

func setRemoteGitURLInTemporaryDirectoryFromNewProject(newRepository *git.Repository, templateURLNew string) error {

	remote := &config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{templateURLNew},
		Fetch: []config.RefSpec{},
	}
	_, errRemote := newRepository.CreateRemote(remote)

	if errRemote != nil {
		return fmt.Errorf("failed to set remote new repository: %v", errRemote)
	}
	return nil
}

func moveTemplateToTempFolder(sourceTemplatePath string, targetTemplatePath string) error {

	// Define the source and destination paths
	src := sourceTemplatePath + "/project-template/golang"
	dest := targetTemplatePath

	// Walk the source directory
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		// Check if the path is the .git directory
		if strings.Contains(path, ".git") {
			// Skip the .git directory and its contents
			return filepath.SkipDir
		}

		// Define the destination path
		destPath := strings.Replace(path, src, dest, 1)

		// Create the destination directory if it doesn't exist
		if info.IsDir() {
			os.MkdirAll(destPath, info.Mode())
			return nil
		}

		// Copy the file to the destination path
		err = copyFile(path, destPath, info.Mode())
		if err != nil {
			return fmt.Errorf("failed to copy %s to %s: %w", path, destPath, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("Error: %s\n", err)
	}
	return nil
}

func replaceDefaultNameStructureOfTheTemplateToNewRepository(templatePathNew string, moduleName string, worktreeNew *git.Worktree) error {

	err := filepath.Walk(templatePathNew, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(path, ".git") {
			return nil // skip the .git directory
		}
		if info.IsDir() {
			return nil
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		content = bytes.ReplaceAll(content, []byte("<module_name>"), []byte(moduleName))
		relPath, err := filepath.Rel(templatePathNew, path)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(path, content, 0666)
		if err != nil {
			return err
		}
		_, err = worktreeNew.Add(relPath)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to replace <module_name> placeholder: %v", err)
	}
	return nil
}

func addChangesInTheGithubOfNewRepository(worktreeNew *git.Worktree) error {
	_, err := worktreeNew.Add(".")
	if err != nil {
		return fmt.Errorf("failed to add files to index: %v", err)
	}
	return nil
}

func commitChangesInTheGithubOfNewRepository(moduleName string, worktreeNew *git.Worktree) error {
	commitMsg := fmt.Sprintf("Initial commit for %s", moduleName)
	_, err := worktreeNew.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Devspace BOT",
			Email: "devspace@devspace.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit files: %v", err)
	}
	return nil
}

func pushChangesInTheGithubOfNewRepository(token string, newRepository *git.Repository) error {
	// Push the changes to the new repository
	err := newRepository.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: &http.BasicAuth{
			Username: "Devspace", // yes, this can be anything except an empty string
			Password: token,
		},
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to push files to repository: %v", err)
	}
	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Sync()
	if err != nil {
		return err
	}

	err = os.Chmod(dst, mode)
	if err != nil {
		return err
	}

	return nil
}
