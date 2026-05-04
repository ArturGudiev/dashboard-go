package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"
	"context"
	"fmt"
	"sort"
)

const taskReportDateTimeLayout = "2006-01-02 15:04:05"

// ReportService builds task completion report trees (parity with Node getTaskReport).
type ReportService struct {
	containerService         *ContainerService
	childContainerRepository *ChildContainerRepository
	tasksRepository          *TasksRepository
}

// NewReportService creates a ReportService.
func NewReportService(
	containerService *ContainerService,
	childContainerRepository *ChildContainerRepository,
	tasksRepository *TasksRepository,
) *ReportService {
	return &ReportService{
		containerService:         containerService,
		childContainerRepository: childContainerRepository,
		tasksRepository:          tasksRepository,
	}
}

// GetTaskReport builds a tree of done tasks under rootTaskID. Returns (nil, nil) when there are no done descendant tasks.
func (s *ReportService) GetTaskReport(ctx context.Context, rootTaskID int) (*models.TaskReportTreeNode, error) {
	_, err := s.tasksRepository.GetTask(ctx, rootTaskID)
	if err != nil {
		return nil, err
	}

	doneTasks, err := s.collectDeepDoneTasks(ctx, rootTaskID)
	if err != nil {
		return nil, err
	}
	if len(doneTasks) == 0 {
		return nil, nil
	}

	rootDesc, err := s.containerService.GetDescription(ctx, schema.ContainerTypeTask, rootTaskID)
	if err != nil {
		return nil, err
	}

	type item struct {
		parentsPath []string
		taskLine    string
	}
	items := make([]item, 0, len(doneTasks))
	for _, t := range doneTasks {
		path, err := s.ancestorDescriptionsBelowRoot(ctx, t.ID, rootTaskID)
		if err != nil {
			return nil, err
		}
		items = append(items, item{
			parentsPath: path,
			taskLine:    taskStringForReport(t),
		})
	}

	first := items[0]
	chain := make([]string, 0, 1+len(first.parentsPath)+1)
	chain = append(chain, *rootDesc)
	chain = append(chain, first.parentsPath...)
	chain = append(chain, first.taskLine)
	tree := treeFromChain(chain, 0)

	for i := 1; i < len(items); i++ {
		appendElementToReportTree(tree, items[i].parentsPath, items[i].taskLine)
	}
	return tree, nil
}

func taskStringForReport(t *ent.Task) string {
	if t.DoneDateTime != nil {
		return fmt.Sprintf("%s --- %s", t.DoneDateTime.Format(taskReportDateTimeLayout), t.Description)
	}
	return t.Description
}

func sortTasksByFinishedDateTime(tasks []*ent.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		a, b := tasks[i].DoneDateTime, tasks[j].DoneDateTime
		if a == nil && b == nil {
			return false
		}
		if a == nil {
			return true
		}
		if b == nil {
			return false
		}
		return a.Before(*b)
	})
}

func (s *ReportService) collectDeepDoneTasks(ctx context.Context, rootTaskID int) ([]*ent.Task, error) {
	var out []*ent.Task
	var dfs func(taskID int) error
	dfs = func(taskID int) error {
		rels, err := s.childContainerRepository.GetChildContainers(ctx, schema.ContainerTypeTask, taskID, schema.ContainerTypeTask)
		if err != nil {
			return err
		}
		for _, rel := range rels {
			t, err := s.tasksRepository.GetTask(ctx, rel.ChildID)
			if err != nil {
				return err
			}
			if t.Done {
				out = append(out, t)
			}
			if err := dfs(rel.ChildID); err != nil {
				return err
			}
		}
		return nil
	}
	if err := dfs(rootTaskID); err != nil {
		return nil, err
	}
	sortTasksByFinishedDateTime(out)
	return out, nil
}

// ancestorDescriptionsBelowRoot returns full-description strings from just below the root task
// down to the parent of leafTaskID (TS getParentsPathInContainers(...).slice(0, -1) using first-parent chain).
func (s *ReportService) ancestorDescriptionsBelowRoot(ctx context.Context, leafTaskID, rootTaskID int) ([]string, error) {
	var bottomUp []models.ContainerDescription
	bottomUp = append(bottomUp, models.ContainerDescription{Type: schema.ContainerTypeTask, ID: leafTaskID})

	for {
		cur := bottomUp[len(bottomUp)-1]
		pt, pid, err := s.containerService.GetParentCommon(ctx, cur.Type, cur.ID)
		if err != nil {
			return nil, fmt.Errorf("walk parents from task %d toward root %d: %w", leafTaskID, rootTaskID, err)
		}
		if *pt == schema.ContainerTypeTask && pid == rootTaskID {
			break
		}
		bottomUp = append(bottomUp, models.ContainerDescription{Type: *pt, ID: pid})
	}

	for i, j := 0, len(bottomUp)-1; i < j; i, j = i+1, j-1 {
		bottomUp[i], bottomUp[j] = bottomUp[j], bottomUp[i]
	}
	if len(bottomUp) <= 1 {
		return nil, nil
	}
	ancestorsOnly := bottomUp[:len(bottomUp)-1]

	out := make([]string, 0, len(ancestorsOnly))
	for _, cd := range ancestorsOnly {
		desc, err := s.containerService.GetDescription(ctx, cd.Type, cd.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, *desc)
	}
	return out, nil
}

func treeFromChain(names []string, depth int) *models.TaskReportTreeNode {
	if len(names) == 0 {
		return nil
	}
	root := &models.TaskReportTreeNode{Name: names[0], Depth: depth}
	cur := root
	for i := 1; i < len(names); i++ {
		child := &models.TaskReportTreeNode{Name: names[i], Depth: cur.Depth + 1}
		cur.Children = []*models.TaskReportTreeNode{child}
		cur = child
	}
	return root
}

func addItemToTreeNode(treeNode *models.TaskReportTreeNode, node *models.TaskReportTreeNode) {
	treeNode.Children = append(treeNode.Children, node)
}

func appendElementToReportTree(treeNode *models.TaskReportTreeNode, parentsPath []string, taskLine string) {
	if len(treeNode.Children) == 0 {
		sub := append(append([]string{}, parentsPath...), taskLine)
		treeNode.Children = []*models.TaskReportTreeNode{treeFromChain(sub, treeNode.Depth+1)}
		return
	}
	lastChild := treeNode.Children[len(treeNode.Children)-1]
	if len(parentsPath) == 0 {
		addItemToTreeNode(treeNode, &models.TaskReportTreeNode{
			Name:  taskLine,
			Depth: treeNode.Depth + 1,
		})
		return
	}
	if lastChild.Name == parentsPath[0] {
		appendElementToReportTree(lastChild, parentsPath[1:], taskLine)
		return
	}
	sub := append(append([]string{}, parentsPath...), taskLine)
	addItemToTreeNode(treeNode, treeFromChain(sub, treeNode.Depth+1))
}
