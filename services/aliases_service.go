package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"
	"arturgudiev/dashboard/repositories"
	"context"
	"errors"
	"fmt"
	"strings"
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
		ID:       alias.ID,
		Type:     alias.Type,
		Alias:    alias.Alias,
		ItemID:   alias.ItemID,
		FilePath: alias.FilePath,
	}
	return AliasModel, nil
}

func (s *AliasesService) ToAliasType(ctx context.Context, containerType schema.ContainerType) (schema.AliasType, error) {
	switch containerType {
	case schema.ContainerTypeTask:
		return schema.AliasTypeTask, nil
	case schema.ContainerTypeProblem:
		return schema.AliasTypeProblem, nil
	case schema.ContainerTypeQuestion:
		return schema.AliasTypeQuestion, nil
	case schema.ContainerTypeStory:
		return schema.AliasTypeStory, nil
	case schema.ContainerTypeEpic:
		return schema.AliasTypeEpic, nil
	case schema.ContainerTypeKnowledgeNode:
		return schema.AliasTypeKnowledgeNode, nil
	case schema.ContainerTypeKnowledgeBit:
		return schema.AliasTypeKnowledgeBit, nil
	case schema.ContainerTypeDefinition:
		return schema.AliasTypeDefinition, nil
	case schema.ContainerTypeAction:
		return schema.AliasTypeAction, nil
	case schema.ContainerTypeState:
		return schema.AliasTypeState, nil
	default:
		return "", errors.New("invalid container type")
	}
}

func (s *AliasesService) GetAliasesByTaskContainer(ctx context.Context, containerType schema.ContainerType, ID int) ([]*models.AliasModel, error) {
	aliasType, err := s.ToAliasType(ctx, containerType)
	if err != nil {
		return nil, err
	}
	aliases, err := s.aliasesRepository.GetAliasesByAliasType(ctx, aliasType, ID)
	if err != nil {
		return nil, err
	}
	aliasModels := make([]*models.AliasModel, len(aliases))
	for i, alias := range aliases {
		aliasModels[i] = &models.AliasModel{
			ID:       alias.ID,
			Type:     alias.Type,
			Alias:    alias.Alias,
			ItemID:   alias.ItemID,
			FilePath: alias.FilePath,
		}
	}
	return aliasModels, nil
}

func (s *AliasesService) GetAliasesByFilePath(ctx context.Context, filePath string) ([]*models.AliasModel, error) {
	// println("Getting aliases by file path: " + filePath)
	// utils.WaitForUserInput()
	aliases, err := s.aliasesRepository.GetAliasesByFilePath(ctx, filePath)
	if err != nil {
		return nil, err
	}
	aliasModels := make([]*models.AliasModel, len(aliases))
	for i, alias := range aliases {
		aliasModels[i] = &models.AliasModel{
			ID:       alias.ID,
			Type:     alias.Type,
			Alias:    alias.Alias,
			ItemID:   alias.ItemID,
			FilePath: alias.FilePath,
		}
	}
	// println("Length: ")
	// println(len(aliasModels))
	// utils.WaitForUserInput()
	return aliasModels, nil
}

func (s *AliasesService) PrintAliases(aliases []*models.AliasModel) {
	if len(aliases) > 0 {
		aliasStrings := make([]string, len(aliases))
		for i, alias := range aliases {
			aliasStrings[i] = alias.Alias
		}

		fmt.Printf("\tAliases: %s\n", strings.Join(aliasStrings, ", "))
	}

}


func (s *AliasesService) AddFileAlias(ctx context.Context, filePath string, alias string) (*ent.Alias, error) {

	aliasEntity, err := s.aliasesRepository.CreateFileAlias(ctx, alias, filePath)
	if err != nil {
		return nil, err
	}
	return aliasEntity, nil
}

func (s *AliasesService) RemoveFileAlias(ctx context.Context, filePath string, alias string) (*ent.Alias, error) {
	aliasEntity, err := s.aliasesRepository.RemoveFileAlias(ctx, alias, filePath)
	if err != nil {
		return nil, err
	}
	return aliasEntity, nil
}


func (s *AliasesService) AddContainerAlias(ctx context.Context, containerType schema.ContainerType, id int, alias string) (*ent.Alias, error) {

	aliasType, err := s.ToAliasType(ctx, containerType)
	if err != nil {
		return nil, err
	}
	aliasEntity, err := s.aliasesRepository.CreateAliasByContainerType(ctx, aliasType, id, alias)
	if err != nil {
		return nil, err
	}
	return aliasEntity, nil
}

func (s *AliasesService) RemoveContainerAlias(ctx context.Context, containerType schema.ContainerType, id int, alias string) (*ent.Alias, error) {
	aliasType, err := s.ToAliasType(ctx, containerType)
	if err != nil {
		return nil, err
	}
	aliasEntity, err := s.aliasesRepository.RemoveAliasFromContainer(ctx, aliasType, id, alias)
	if err != nil {
		return nil, err
	}
	return aliasEntity, nil
}