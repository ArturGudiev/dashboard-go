package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"
	"context"
)

type StoriesRepository struct {
	client *ent.Client
}

func NewStoriesRepository(client *ent.Client) *StoriesRepository {
	return &StoriesRepository{client: client}
}

func (r *StoriesRepository) GetStory(ctx context.Context, ID int) (*ent.Story, error) {
	story, err := r.client.Story.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	return story, nil
}

func (r *StoriesRepository) AddStory(ctx context.Context, description string, tags []string, notes string) (*ent.Story, error) {
	story, err := r.client.Story.Create().SetDescription(description).SetTags(tags).SetNotes(notes).Save(ctx)

	if err != nil {
		return nil, err
	}
	return story, nil
}

func (r *StoriesRepository) UpdateStory(ctx context.Context, story models.StoryPartial) error {
	updateBuilder := r.client.Story.UpdateOneID(story.ID)

	if story.Description != nil {
		updateBuilder = updateBuilder.SetDescription(*story.Description)
	}

	if story.Notes != nil {
		updateBuilder = updateBuilder.SetNotes(*story.Notes)
	}

	if story.Tags != nil {
		updateBuilder = updateBuilder.SetTags(*story.Tags)
	}

	if story.Closed != nil {
		updateBuilder = updateBuilder.SetClosed(*story.Closed)
	}

	if story.DoneDateTime != nil {
		updateBuilder = updateBuilder.SetDoneDateTime(*story.DoneDateTime)
	}

	_, err := updateBuilder.Save(ctx)
	return err
}
