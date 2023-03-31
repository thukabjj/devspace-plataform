package handler

import (
	"devspace-backend/dto"
	"devspace-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateRepositoryHandler(c *gin.Context) {
	// TODO: Implement the logic to create repository, pipeline, and index
	var projectRepositoryDTO dto.ProjectRepositoryDTO

	if err := c.ShouldBindJSON(&projectRepositoryDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	service.CreateRepository(projectRepositoryDTO)

	c.JSON(http.StatusOK, gin.H{
		"message": "Repository created successfully!",
	})
}
