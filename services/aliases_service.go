package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/repositories"
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
