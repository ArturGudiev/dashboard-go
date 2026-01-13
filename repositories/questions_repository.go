package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/models"
	"context"
)

type QuestionsRepository struct {
	client *ent.Client
}

func NewQuestionsRepository(client *ent.Client) *QuestionsRepository {
	return &QuestionsRepository{client: client}
}

func (r *QuestionsRepository) AddAnswer(ctx context.Context, ID int, answer string) error {
	updateBuilder := r.client.Question.UpdateOneID(ID).
		SetAnswer(answer)

	_, err := updateBuilder.Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *QuestionsRepository) GetQuestion(ctx context.Context, ID int) (*ent.Question, error) {
	question, err := r.client.Question.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	return question, nil
}

func (r *QuestionsRepository) AddQuestion(ctx context.Context, description string, tags []string, notes string) (*ent.Question, error) {
	question, err := r.client.Question.Create().SetDescription(description).SetTags(tags).SetNotes(notes).Save(ctx)

	if err != nil {
		return nil, err
	}
	return question, nil
}

func (r *QuestionsRepository) UpdateQuestion(ctx context.Context, question models.QuestionPartial) error {
	updateBuilder := r.client.Question.UpdateOneID(question.ID)

	if question.Description != nil {
		updateBuilder = updateBuilder.SetDescription(*question.Description)
	}

	if question.Notes != nil {
		updateBuilder = updateBuilder.SetNotes(*question.Notes)
	}

	if question.Tags != nil {
		updateBuilder = updateBuilder.SetTags(*question.Tags)
	}

	if question.Answer != nil {
		updateBuilder = updateBuilder.SetAnswer(*question.Answer)
	}

	if question.DoneDateTime != nil {
		updateBuilder = updateBuilder.SetDoneDateTime(*question.DoneDateTime)
	}

	_, err := updateBuilder.Save(ctx)
	return err
}
