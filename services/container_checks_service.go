package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"
	"arturgudiev/dashboard/repositories"
	"context"
)

type ContainerChecksService struct {
	repo *repositories.ContainerChecksRepository
}

func NewContainerChecksService(repo *repositories.ContainerChecksRepository) *ContainerChecksService {
	return &ContainerChecksService{repo: repo}
}

func (s *ContainerChecksService) AddCheck(
	ctx context.Context,
	description string,
	containerType schema.ContainerType,
	containerID int,
) (*models.ContainerCheck, error) {
	check, err := s.repo.AddCheck(ctx, description, containerType, containerID)
	if err != nil {
		return nil, err
	}
	return toContainerCheck(check), nil
}

func (s *ContainerChecksService) RemoveCheck(ctx context.Context, id int) error {
	return s.repo.RemoveCheck(ctx, id)
}

func (s *ContainerChecksService) UpdateCheck(
	ctx context.Context,
	id int,
	description string,
) (*models.ContainerCheck, error) {
	check, err := s.repo.UpdateCheck(ctx, id, description)
	if err != nil {
		return nil, err
	}
	return toContainerCheck(check), nil
}

func (s *ContainerChecksService) GetChecksByContainer(
	ctx context.Context,
	containerType schema.ContainerType,
	containerID int,
) ([]models.ContainerCheck, error) {
	checks, err := s.repo.GetChecksByContainer(ctx, containerType, containerID)
	if err != nil {
		return nil, err
	}
	return toContainerChecks(checks), nil
}

func toContainerCheck(check *ent.ContainerCheck) *models.ContainerCheck {
	if check == nil {
		return nil
	}
	return &models.ContainerCheck{
		ID:            check.ID,
		Description:   check.Description,
		ContainerType: check.ContainerType,
		ContainerID:   check.ContainerID,
	}
}

func toContainerChecks(checks []*ent.ContainerCheck) []models.ContainerCheck {
	if len(checks) == 0 {
		return []models.ContainerCheck{}
	}
	result := make([]models.ContainerCheck, len(checks))
	for i, check := range checks {
		result[i] = *toContainerCheck(check)
	}
	return result
}
