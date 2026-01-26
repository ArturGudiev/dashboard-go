package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/containerchild"
	"arturgudiev/dashboard/ent/schema"
	models "arturgudiev/dashboard/models"
	"context"
)

type ChildContainerRepository struct {
	client *ent.Client
}

func NewChildContainerRepository(client *ent.Client) *ChildContainerRepository {
	return &ChildContainerRepository{client: client}
}

func (s *ChildContainerRepository) GetChildContainers(ctx context.Context, parentType schema.ContainerType, parentID int, childrenType schema.ContainerType) ([]*ent.ContainerChild, error) {
	childRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ParentTypeEQ(parentType),
			containerchild.ParentID(parentID),
			containerchild.ChildTypeEQ(childrenType),
		).
		Order(containerchild.ByChildOrder()).
		All(ctx)

	if err == nil {
		return childRelations, nil
	}
	return nil, err
}

func (s *ChildContainerRepository) GetParentContainers(ctx context.Context, childType schema.ContainerType, childID int) ([]models.ContainerDescription, error) {
	parentContainers := []models.ContainerDescription{}

	parentRelations, err := s.client.ContainerChild.Query().
		Where(
			containerchild.ChildTypeEQ(childType),
			containerchild.ChildID(childID),
		).
		Order(containerchild.ByParentOrder()).
		All(ctx)

	if err != nil {
		return nil, err
	}

	for _, t := range parentRelations {
		parentContainer := models.ContainerDescription{ID: t.ParentID, Type: t.ParentType}
		parentContainers = append(parentContainers, parentContainer)
	}

	return parentContainers, nil
}

func (s *ChildContainerRepository) AddConnection(ctx context.Context, parentType schema.ContainerType, parentID int,
	childType schema.ContainerType, childID int) (*ent.ContainerChild, error) {

	maxChildOrders, _ := s.client.ContainerChild.Query().
		Where(
			containerchild.ChildTypeEQ(childType),
			containerchild.ParentIDEQ(parentID),
			containerchild.ParentTypeEQ(parentType),
		).
		Aggregate(ent.Max(containerchild.FieldChildOrder)).
		Ints(ctx)

	maxParentOrders, _ := s.client.ContainerChild.Query().
		Where(
			containerchild.ChildTypeEQ(childType),
			containerchild.ChildID(childID),
			containerchild.ParentTypeEQ(parentType),
		).
		Aggregate(ent.Max(containerchild.FieldChildOrder)).
		Ints(ctx)

	maxChildOrder := 0
	if len(maxChildOrders) > 0 {
		maxChildOrder = maxChildOrders[0]
	}

	maxParentOrder := 0
	if len(maxParentOrders) > 0 {
		maxParentOrder = maxParentOrders[0]
	}

	nextChildOrder := maxChildOrder + 1
	nextParentOrder := maxParentOrder + 1

	newRelation, err := s.client.ContainerChild.Create().
		SetParentID(parentID).
		SetParentType(parentType).
		SetChildID(childID).
		SetChildType(childType).
		SetChildOrder(nextChildOrder).
		SetParentOrder(nextParentOrder).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return newRelation, nil
}
