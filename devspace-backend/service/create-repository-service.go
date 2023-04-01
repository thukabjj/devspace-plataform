package service

import (
	"devspace-backend/dto"
)

func CreateRepository(productRepositoryDTO dto.ProjectRepositoryDTO) error {
	// Create the GitHUB Repository
	return CreateGithubRepository(productRepositoryDTO)
	// Create Jenkins Pipeline
	//createJenkinsPipeline(productRepositoryDTO)
	// Create Elasticsearch Index
}
