package dto

type ProjectRepositoryDTO struct {
	// Name is required
	Name        string `json:"name" binding:"required"`
	Language    string `json:"language" binding:"required"`
	Description string `json:"description" binding:"required"`
}
