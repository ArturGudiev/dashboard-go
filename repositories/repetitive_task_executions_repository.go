package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/repetitivetaskexecution"
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
)

type RepetitiveTaskExecutionsRepository struct {
	client *ent.Client
}

func NewRepetitiveTaskExecutionsRepository(client *ent.Client) *RepetitiveTaskExecutionsRepository {
	return &RepetitiveTaskExecutionsRepository{client: client}
}

// AddRepetitiveTaskExecution records an execution for the given repetitive task at the current time.
func (r *RepetitiveTaskExecutionsRepository) AddRepetitiveTaskExecution(ctx context.Context, repetitiveTaskID int) (*ent.RepetitiveTaskExecution, error) {
	return r.client.RepetitiveTaskExecution.Create().
		SetRepetitiveTaskID(repetitiveTaskID).
		SetExecutionDate(time.Now()).
		Save(ctx)
}

// GetRepetitiveTaskExecutions returns all executions for a repetitive task, newest first.
func (r *RepetitiveTaskExecutionsRepository) GetRepetitiveTaskExecutions(ctx context.Context, repetitiveTaskID int) ([]*ent.RepetitiveTaskExecution, error) {
	return r.client.RepetitiveTaskExecution.Query().
		Where(repetitivetaskexecution.RepetitiveTaskIDEQ(repetitiveTaskID)).
		Order(repetitivetaskexecution.ByExecutionDate(sql.OrderDesc())).
		All(ctx)
}
