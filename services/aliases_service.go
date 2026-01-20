package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"
	"arturgudiev/dashboard/repositories"
	"context"
)

type AliasesService struct {
	client                   *ent.Client
	containerService         *ContainerService
	aliasesRepository        *repositories.AliasesRepository
	childContainerRepository *ChildContainerRepository
}

func NewAliasesService(client *ent.Client, containerService *ContainerService, aliasesRepository *repositories.AliasesRepository, childContainerRepository *ChildContainerRepository) *AliasesService {
	return &AliasesService{client: client, containerService: containerService, aliasesRepository: aliasesRepository, childContainerRepository: childContainerRepository}
}

func (s *AliasesService) GetAlias(ctx context.Context, aliasString string) (*models.AliasModel, error) {
	alias, errProblem := s.aliasesRepository.GetAliasByAliasString(ctx, aliasString)
	if errProblem != nil {
		return nil, errProblem
	}

	AliasModel := &models.AliasModel{
		ID:     alias.ID,
		Type:   alias.Type,
		Alias:  alias.Alias,
		ItemID: alias.ItemID,
	}
	return AliasModel, nil
}
