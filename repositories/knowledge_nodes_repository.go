package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"
	"context"
)

type KnowledgeNodesRepository struct {
	client *ent.Client
}

func NewKnowledgeNodesRepository(client *ent.Client) *KnowledgeNodesRepository {
	return &KnowledgeNodesRepository{client: client}
}

func (r *KnowledgeNodesRepository) GetAllKnowledgeNodes(ctx context.Context) ([]*ent.KnowledgeNode, error) {
	knowledgeNodes, err := r.client.KnowledgeNode.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	return knowledgeNodes, nil
}

func (r *KnowledgeNodesRepository) GetKnowledgeNode(ctx context.Context, ID int) (*ent.KnowledgeNode, error) {
	knowledgeNode, err := r.client.KnowledgeNode.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	return knowledgeNode, nil
}

func (r *KnowledgeNodesRepository) AddKnowledgeNode(ctx context.Context, name string, tags []string, notes string) (*ent.KnowledgeNode, error) {
	knowledgeNode, err := r.client.KnowledgeNode.Create().SetName(name).SetTags(tags).SetNotes(notes).Save(ctx)

	if err != nil {
		return nil, err
	}
	return knowledgeNode, nil
}

func (r *KnowledgeNodesRepository) UpdateKnowledgeNode(ctx context.Context, knowledgeNode models.KnowledgeNodePartial) error {
	updateBuilder := r.client.KnowledgeNode.UpdateOneID(knowledgeNode.ID)

	if knowledgeNode.Name != nil {
		updateBuilder = updateBuilder.SetName(*knowledgeNode.Name)
	}

	if knowledgeNode.Notes != nil {
		updateBuilder = updateBuilder.SetNotes(*knowledgeNode.Notes)
	}

	if knowledgeNode.Tags != nil {
		updateBuilder = updateBuilder.SetTags(*knowledgeNode.Tags)
	}

	_, err := updateBuilder.Save(ctx)
	return err
}
