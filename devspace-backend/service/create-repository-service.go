package service

import (
	"context"
	"devspace-backend/dto"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

func CreateRepository(productRepositoryDTO dto.ProjectRepositoryDTO) {
	// Create the GitHUB Repository
	createGithubRepository(productRepositoryDTO)
	// Create Jenkins Pipeline
	//createJenkinsPipeline(productRepositoryDTO)
	// Create Elasticsearch Index
}

func createGithubRepository(productRepositoryDTO dto.ProjectRepositoryDTO) {

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("GITHUB_TOKEN environment variable is not set")
		os.Exit(1)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repo := &github.Repository{
		Name:        github.String(productRepositoryDTO.Name),
		Description: github.String(productRepositoryDTO.Description),
		Language:    github.String(productRepositoryDTO.Language),
		Private:     github.Bool(true),
	}

	createdRepo, _, err := client.Repositories.Create(ctx, "", repo)
	if err != nil {
		fmt.Printf("Failed to create repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created repository: %s\n", createdRepo.GetName())
	/*
		// Copy template to new repository
		projectPath := path.Join("/mnt/wsl", "Ubuntu", "home", "arthur", "dev", "devspace-plataform", productRepositoryDTO.Name)
		templatePath := path.Join("/mnt/wsl", "Ubuntu", "home", "arthur", "dev", "devspace-plataform", "project-template", "golang")

		// Create the new project directory
		if err := os.MkdirAll(projectPath, 0755); err != nil {
			fmt.Printf("Failed to create new project directory: %v\n", err)
			os.Exit(1)
		}

		// Copy files from template to new project directory
		if err := filepath.Walk(templatePath, func(filePath string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get relative file path within template directory
			relPath, err := filepath.Rel(templatePath, filePath)
			if err != nil {
				return err
			}

			// Skip directories
			if fileInfo.IsDir() {
				return nil
			}

			// Read template file contents
			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			// Replace module name in go.mod file
			if relPath == "go.mod" {
				content = []byte(fmt.Sprintf(string(content), productRepositoryDTO.Name))
			}

			// Write contents to new project file
			destPath := filepath.Join(projectPath, relPath)
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}

			if err := ioutil.WriteFile(destPath, content, 0644); err != nil {
				return err
			}

			return nil
		}); err != nil {
			fmt.Printf("Failed to copy template to new repository: %v\n", err)
			os.Exit(1)
		}

		// Initialize git repository in new project directory
		repo2, err := git.PlainInit(projectPath, false)
		if err != nil {
			fmt.Printf("Failed to initialize git repository in new project directory: %v\n", err)
			os.Exit(1)
		}

		// Add files to git and commit changes
		if err := commitAndPushChanges(repo2, "Initial commit"); err != nil {
			fmt.Printf("Failed to commit and push changes: %v\n", err)
			os.Exit(1)
		}
	*/
	fmt.Println("Successfully created new repository and pushed template to it!")
}

// copyTemplate copies the contents of the template directory to the new repository directory
func copyTemplate(templatePath, projectPath string) error {
	cmd := exec.Command("cp", "-r", templatePath+"/.", projectPath)
	return cmd.Run()
}

// replaceModuleName replaces the module name in the go.mod file with the provided name
func replaceModuleName(modFilePath, moduleName string) error {
	moduleNamePattern := regexp.MustCompile(`module .*`)
	moduleNameLine := fmt.Sprintf("module %s", moduleName)

	input, err := ioutil.ReadFile(modFilePath)
	if err != nil {
		return err
	}

	output := moduleNamePattern.ReplaceAll(input, []byte(moduleNameLine))

	return ioutil.WriteFile(modFilePath, output, 0666)
}

func commitAndPushChanges(repo *git.Repository, commitMessage string) error {

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}

	_, err = worktree.Add(".")
	if err != nil {
		return fmt.Errorf("failed to stage changes: %v", err)
	}

	commit, err := worktree.Commit(commitMessage, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "devspace-backend",
			Email: "devspace@backend.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit changes: %v", err)
	}

	auth := &http.BasicAuth{
		Username: "devspace-backend",
		Password: "devspace",
	}

	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
		Progress:   os.Stdout,
		RefSpecs:   []config.RefSpec{"refs/heads/*:refs/heads/*"},
	})
	if err != nil {
		return fmt.Errorf("failed to push changes to remote: %v", err)
	}

	fmt.Printf("pushed commit %s\n", commit)

	return nil
}

func createJenkinsPipeline(productRepositoryDTO dto.ProjectRepositoryDTO) error {

	jenkinsToken := os.Getenv("JENKINS_API_TOKEN")
	if jenkinsToken == "" {
		fmt.Println("JENKINS_API_TOKEN environment variable is not set")
		os.Exit(1)
	}

	// Set up Jenkins client
	jenkinsURL := "http://localhost:8080"
	jenkinsUser := "jenkins"
	jenkins := gojenkins.CreateJenkins(nil, jenkinsURL, jenkinsUser, jenkinsToken)
	githubRepoName := productRepositoryDTO.Name

	// Create Jenkins pipeline
	jobName := githubRepoName + "-pipeline"
	jobConfig := `<flow-definition plugin="workflow-job@2.40">
    <actions/>
    <description>Simple pipeline for ` + githubRepoName + `</description>
    <keepDependencies>false</keepDependencies>
    <properties/>
    <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps@2.80">
		<script>pipeline {
		agent any

		stages {
			stage('Build') {
				steps {
					sh 'echo "Building..."'
				}
			}
		}
	}</script>
		<sandbox>true</sandbox>
		</definition>
		<triggers/>
		<disabled>false</disabled>
	</flow-definition>`

	ctx := context.Background()

	if _, err := jenkins.CreateJob(ctx, jobConfig, jobName); err != nil {
		return err
	}

	// Log successful creation
	log.Printf("Successfully created Github repository '%s' and Jenkins pipeline job '%s'", githubRepoName, jobName)

	return nil
}
