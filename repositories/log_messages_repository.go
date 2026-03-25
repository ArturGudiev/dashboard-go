package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/logmessage"
	"arturgudiev/dashboard/ent/schema"
	"context"
	"time"
)

type LogMessagesRepository struct {
	client *ent.Client
}

func NewLogMessagesRepository(client *ent.Client) *LogMessagesRepository {
	return &LogMessagesRepository{client: client}
}

func (r *LogMessagesRepository) AddLogMessage(ctx context.Context, description string, containerType *schema.ContainerType, containerID *int) (*ent.LogMessage, error) {
	logMessage, err := r.client.LogMessage.Create().
		SetDescription(description).
		SetNillableContainerType(containerType).
		SetNillableContainerID(containerID).
		SetCreated(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return logMessage, nil
}

func (r *LogMessagesRepository) GetLogMessage(ctx context.Context, ID int) (*ent.LogMessage, error) {
	logMessage, err := r.client.LogMessage.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	return logMessage, nil
}

func (r *LogMessagesRepository) GetLogMessages(ctx context.Context, perPage int, page int) ([]*ent.LogMessage, *int, error) {
	logMessages, err := r.client.LogMessage.Query().
		Offset(page * perPage).
		Limit(perPage).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}
	total, err := r.client.LogMessage.Query().Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	return logMessages, &total, nil
}

func (r *LogMessagesRepository) GetContainerLogMessages(ctx context.Context, containerType schema.ContainerType, containerID int, perPage int, page int) ([]*ent.LogMessage, *int, error) {
	logMessages, err := r.client.LogMessage.Query().
		Where(logmessage.ContainerTypeEQ(containerType), logmessage.ContainerIDEQ(containerID)).
		Offset(page * perPage).
		Limit(perPage).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}
	total, err := r.client.LogMessage.Query().Where(logmessage.ContainerTypeEQ(containerType), logmessage.ContainerIDEQ(containerID)).Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	return logMessages, &total, nil
}

// func (r *LogMessagesRepository) AddQuestion(ctx context.Context, description string, tags []string, notes string) (*ent.Question, error) {
// 	question, err := r.client.Question.Create().SetDescription(description).SetTags(tags).SetNotes(notes).Save(ctx)

// 	if err != nil {
// 		return nil, err
// 	}
// 	return question, nil
// }

// func (r *LogMessagesRepository) UpdateQuestion(ctx context.Context, question models.QuestionPartial) error {
// 	updateBuilder := r.client.Question.UpdateOneID(question.ID)

// 	if question.Description != nil {
// 		updateBuilder = updateBuilder.SetDescription(*question.Description)
// 	}

// 	if question.Notes != nil {
// 		updateBuilder = updateBuilder.SetNotes(*question.Notes)
// 	}

// 	if question.Tags != nil {
// 		updateBuilder = updateBuilder.SetTags(*question.Tags)
// 	}

// 	if question.Answer != nil {
// 		updateBuilder = updateBuilder.SetAnswer(*question.Answer)
// 	}

// 	if question.DoneDateTime != nil {
// 		updateBuilder = updateBuilder.SetDoneDateTime(*question.DoneDateTime)
// 	}

// 	_, err := updateBuilder.Save(ctx)
// 	return err
// }
