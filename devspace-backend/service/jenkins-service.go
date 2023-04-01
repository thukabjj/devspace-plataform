package service

import (
	"context"
	"devspace-backend/dto"
	"fmt"
	"log"
	"os"

	"github.com/bndr/gojenkins"
)

func CreateJenkinsPipeline(productRepositoryDTO dto.ProjectRepositoryDTO) error {

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
