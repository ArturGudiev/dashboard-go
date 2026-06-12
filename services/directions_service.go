package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"
	"arturgudiev/dashboard/repositories"
	"context"
	"errors"
	"fmt"
	"sort"
	"time"
)

type DirectionsService struct {
	directionsRepository           *repositories.DirectionsRepository
	directionSubmissionsRepository *repositories.DirectionSubmissionsRepository
	longTaskSubmissionsRepository  *repositories.LongTaskSubmissionsRepository
	longTasksRepository            *repositories.LongTasksRepository
	containerService               *ContainerService
	childContainerRepository       *ChildContainerRepository
}

func NewDirectionsService(
	directionsRepository *repositories.DirectionsRepository,
	directionSubmissionsRepository *repositories.DirectionSubmissionsRepository,
	longTaskSubmissionsRepository *repositories.LongTaskSubmissionsRepository,
	longTasksRepository *repositories.LongTasksRepository,
	containerService *ContainerService,
	childContainerRepository *ChildContainerRepository,
) *DirectionsService {
	return &DirectionsService{
		directionsRepository:           directionsRepository,
		directionSubmissionsRepository: directionSubmissionsRepository,
		longTaskSubmissionsRepository:  longTaskSubmissionsRepository,
		longTasksRepository:            longTasksRepository,
		containerService:               containerService,
		childContainerRepository:       childContainerRepository,
	}
}

func (s *DirectionsService) GetDirections(ctx context.Context, open *bool) ([]*ent.Direction, error) {
	return s.directionsRepository.GetDirections(ctx, open)
}

func (s *DirectionsService) GetDirectionById(ctx context.Context, id int) (*models.DirectionFull, error) {
	return s.getDirectionFull(ctx, id)
}

func (s *DirectionsService) getDirectionFull(ctx context.Context, id int) (*models.DirectionFull, error) {
	direction, err := s.directionsRepository.GetDirection(ctx, id)
	if err != nil {
		return nil, err
	}

	subtasks, errSubtasks := s.containerService.GetOpenSubtasksIDs(ctx, schema.ContainerTypeDirection, id)
	subproblems, errSubproblems := s.containerService.GetOpenProblemsIDs(ctx, schema.ContainerTypeDirection, id)
	subquestions, errSubquestions := s.containerService.GetOpenQuestionsIDs(ctx, schema.ContainerTypeDirection, id)
	substories, errSubstories := s.containerService.GetOpenStoriesIDs(ctx, schema.ContainerTypeDirection, id)
	subdirections, errSubdirections := s.containerService.GetOpenDirectionsIDs(ctx, schema.ContainerTypeDirection, id)
	longTasks, errLongTasks := s.containerService.GetOpenLongTasksIDs(ctx, schema.ContainerTypeDirection, id)
	parentContainers, errParents := s.childContainerRepository.GetParentContainers(ctx, schema.ContainerTypeDirection, id)

	if errSubtasks != nil || errSubproblems != nil || errSubquestions != nil ||
		errSubstories != nil || errSubdirections != nil || errLongTasks != nil || errParents != nil {
		return nil, errors.New("direction not found")
	}

	return &models.DirectionFull{
		ID:               id,
		Description:      direction.Description,
		Tags:             direction.Tags,
		Notes:            direction.Notes,
		Closed:           direction.Closed,
		Tasks:            subtasks,
		Problems:         subproblems,
		Questions:        subquestions,
		Stories:          substories,
		Directions:       subdirections,
		LongTasks:        longTasks,
		ParentContainers: parentContainers,
	}, nil
}

func (s *DirectionsService) AddDirection(
	ctx context.Context,
	direction models.DirectionShort,
	parent *models.ContainerDescription,
) (*models.DirectionFull, error) {
	newDirection, err := s.directionsRepository.AddDirection(ctx, direction.Description, direction.Tags, direction.Notes)
	if err != nil {
		return nil, err
	}
	if parent != nil {
		_, err := s.childContainerRepository.AddConnection(
			ctx,
			parent.Type,
			parent.ID,
			schema.ContainerTypeDirection,
			newDirection.ID,
		)
		if err != nil {
			return nil, err
		}
	}
	return s.getDirectionFull(ctx, newDirection.ID)
}

func (s *DirectionsService) UpdateDirection(ctx context.Context, partial models.DirectionPartial) (*models.DirectionFull, error) {
	if err := s.directionsRepository.UpdateDirection(ctx, partial); err != nil {
		return nil, err
	}
	return s.getDirectionFull(ctx, partial.ID)
}

func (s *DirectionsService) AddDirectionSubmission(
	ctx context.Context,
	directionID int,
	text *string,
) (*ent.DirectionSubmission, error) {
	if _, err := s.directionsRepository.GetDirection(ctx, directionID); err != nil {
		return nil, err
	}
	return s.directionSubmissionsRepository.AddDirectionSubmission(ctx, directionID, text)
}

func (s *DirectionsService) GetDirectionSubmissions(
	ctx context.Context,
	directionID int,
) ([]*ent.DirectionSubmission, error) {
	return s.directionSubmissionsRepository.GetDirectionSubmissions(ctx, directionID)
}

func (s *DirectionsService) GetDirectionStats(ctx context.Context, directionID int) ([]models.DirectionStatsEntry, error) {
	if _, err := s.directionsRepository.GetDirection(ctx, directionID); err != nil {
		return nil, err
	}

	type datedEntry struct {
		date time.Time
		text string
	}
	var entries []datedEntry

	directionSubmissions, err := s.directionSubmissionsRepository.GetDirectionSubmissions(ctx, directionID)
	if err != nil {
		return nil, err
	}
	for _, submission := range directionSubmissions {
		text := "—"
		if submission.Text != nil && *submission.Text != "" {
			text = *submission.Text
		}
		entries = append(entries, datedEntry{date: submission.ExecutionDate, text: text})
	}

	taskIDs, longTaskIDs, err := s.containerService.CollectDescendantTaskAndLongTaskIDs(
		ctx,
		schema.ContainerTypeDirection,
		directionID,
	)
	if err != nil {
		return nil, err
	}

	if len(taskIDs) > 0 {
		doneTasks, err := s.containerService.GetDoneTasksByIDs(ctx, taskIDs)
		if err != nil {
			return nil, err
		}

		doneByDay := map[string]int{}
		dayTimes := map[string]time.Time{}
		for _, doneTask := range doneTasks {
			if doneTask.DoneDateTime == nil {
				continue
			}
			dayKey := doneTask.DoneDateTime.Format("2006-01-02")
			doneByDay[dayKey]++
			if _, ok := dayTimes[dayKey]; !ok {
				dayTimes[dayKey] = time.Date(
					doneTask.DoneDateTime.Year(),
					doneTask.DoneDateTime.Month(),
					doneTask.DoneDateTime.Day(),
					0, 0, 0, 0,
					doneTask.DoneDateTime.Location(),
				)
			}
		}
		for dayKey, count := range doneByDay {
			label := "tasks were done"
			if count == 1 {
				label = "task was done"
			}
			entries = append(entries, datedEntry{
				date: dayTimes[dayKey],
				text: fmt.Sprintf("%d %s", count, label),
			})
		}
	}

	for _, longTaskID := range longTaskIDs {
		longTask, err := s.longTasksRepository.GetLongTask(ctx, longTaskID)
		if err != nil {
			continue
		}
		submissions, err := s.longTaskSubmissionsRepository.GetLongTaskSubmissions(ctx, longTaskID)
		if err != nil {
			continue
		}
		for _, submission := range submissions {
			entries = append(entries, datedEntry{
				date: submission.ExecutionDate,
				text: formatLongTaskSubmissionStatLine(longTask, submission),
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].date.Equal(entries[j].date) {
			return entries[i].text < entries[j].text
		}
		return entries[i].date.After(entries[j].date)
	})

	result := make([]models.DirectionStatsEntry, len(entries))
	for i, entry := range entries {
		result[i] = models.DirectionStatsEntry{
			Date: entry.date.Format(time.RFC3339),
			Text: entry.text,
		}
	}
	return result, nil
}

func formatLongTaskSubmissionStatLine(longTask *ent.LongTask, submission *ent.LongTaskSubmission) string {
	submissionText := longTaskSubmissionText(submission)
	return fmt.Sprintf("LongTask %d %s has submission %s", longTask.ID, longTask.Description, submissionText)
}

func longTaskSubmissionText(submission *ent.LongTaskSubmission) string {
	if submission.ProgressRaw != nil && *submission.ProgressRaw != "" {
		return *submission.ProgressRaw
	}
	if submission.ProgressToSet != nil {
		return fmt.Sprintf("set to %v", *submission.ProgressToSet)
	}
	if submission.ProgressToAdd != nil {
		return fmt.Sprintf("+%v", *submission.ProgressToAdd)
	}
	if submission.Comments != nil && *submission.Comments != "" {
		return *submission.Comments
	}
	return "—"
}
