package services

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/models"
	"arturgudiev/dashboard/repositories"
	"arturgudiev/dashboard/scripting"
	"context"
	"fmt"
	"strconv"
	"strings"
)

type ScriptsService struct {
	repo            *repositories.ScriptsRepository
	taskService     *TaskService
	problemService  *ProblemService
	variablesRepo   *repositories.ContainerVariablesRepository
}

func NewScriptsService(
	repo *repositories.ScriptsRepository,
	taskService *TaskService,
	problemService *ProblemService,
	variablesRepo *repositories.ContainerVariablesRepository,
) *ScriptsService {
	return &ScriptsService{
		repo:           repo,
		taskService:    taskService,
		problemService: problemService,
		variablesRepo:  variablesRepo,
	}
}

func (s *ScriptsService) Create(ctx context.Context, short models.ScriptShort) (*models.ScriptFull, error) {
	if err := validateScriptMeta(short.Name, short.Code, short.Params); err != nil {
		return nil, err
	}
	var containerType *schema.ContainerType
	var containerID *int
	if short.Container != nil {
		if short.Container.ID <= 0 || short.Container.Type == "" {
			return nil, fmt.Errorf("container id and type are required for local scripts")
		}
		ct := short.Container.Type
		id := short.Container.ID
		containerType = &ct
		containerID = &id
	}
	created, err := s.repo.Create(
		ctx,
		short.Name,
		short.Code,
		short.Description,
		models.ToSchemaParams(short.Params),
		containerType,
		containerID,
	)
	if err != nil {
		return nil, err
	}
	return toScriptFull(created), nil
}

func (s *ScriptsService) Get(ctx context.Context, id int) (*models.ScriptFull, error) {
	script, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return toScriptFull(script), nil
}

func (s *ScriptsService) List(ctx context.Context, filter models.ScriptListFilter) ([]models.ScriptListItem, error) {
	scope := strings.ToLower(strings.TrimSpace(filter.Scope))
	if scope == "" {
		if filter.ContainerID > 0 && filter.ContainerType != "" {
			scope = "all"
		} else {
			scope = "global"
		}
	}
	if scope == "local" && (filter.ContainerID <= 0 || filter.ContainerType == "") {
		return nil, fmt.Errorf("containerType and containerId are required for local scope")
	}
	if scope == "all" && (filter.ContainerID <= 0 || filter.ContainerType == "") {
		scope = "global"
	}

	scripts, err := s.repo.List(ctx, filter.Query, scope, filter.ContainerType, filter.ContainerID)
	if err != nil {
		return nil, err
	}
	items := make([]models.ScriptListItem, len(scripts))
	for i, script := range scripts {
		items[i] = toScriptListItem(script)
	}
	return items, nil
}

func (s *ScriptsService) Update(ctx context.Context, id int, partial models.ScriptPartial) (*models.ScriptFull, error) {
	var schemaParams *[]schema.ScriptParam
	if partial.Params != nil {
		if err := validateParams(*partial.Params); err != nil {
			return nil, err
		}
		converted := models.ToSchemaParams(*partial.Params)
		schemaParams = &converted
	}
	if partial.Name != nil && strings.TrimSpace(*partial.Name) == "" {
		return nil, fmt.Errorf("name must not be empty")
	}
	if partial.Code != nil {
		if strings.TrimSpace(*partial.Code) == "" {
			return nil, fmt.Errorf("code must not be empty")
		}
		if err := scripting.Validate(*partial.Code); err != nil {
			return nil, fmt.Errorf("syntax error: %w", err)
		}
	}
	updated, err := s.repo.Update(ctx, id, partial.Name, partial.Code, partial.Description, schemaParams)
	if err != nil {
		return nil, err
	}
	return toScriptFull(updated), nil
}

func (s *ScriptsService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *ScriptsService) Validate(code string) models.ScriptValidateResponse {
	if strings.TrimSpace(code) == "" {
		return models.ScriptValidateResponse{OK: false, Error: "code must not be empty"}
	}
	if err := scripting.Validate(code); err != nil {
		return models.ScriptValidateResponse{OK: false, Error: err.Error()}
	}
	return models.ScriptValidateResponse{OK: true}
}

func (s *ScriptsService) Execute(
	ctx context.Context,
	scriptID int,
	container models.ContainerDescription,
	params map[string]any,
) (*models.ScriptRunResponse, error) {
	script, err := s.repo.Get(ctx, scriptID)
	if err != nil {
		return nil, err
	}

	merged, err := mergeAndValidateParams(models.FromSchemaParams(script.Params), params)
	if err != nil {
		return &models.ScriptRunResponse{OK: false, Error: err.Error(), Created: emptyCreated()}, nil
	}

	host := &scriptHost{
		ctx:            ctx,
		container:      container,
		taskService:    s.taskService,
		problemService: s.problemService,
		variablesRepo:  s.variablesRepo,
	}

	result, err := scripting.Run(ctx, script.Code, merged, host)
	if err != nil {
		return &models.ScriptRunResponse{OK: false, Error: err.Error(), Created: emptyCreated()}, nil
	}

	created := emptyCreated()
	if result != nil {
		created.Tasks = result.CreatedTasks
		created.Problems = result.CreatedProblems
		if created.Tasks == nil {
			created.Tasks = []int{}
		}
		if created.Problems == nil {
			created.Problems = []int{}
		}
	}
	return &models.ScriptRunResponse{OK: true, Created: created}, nil
}

type scriptHost struct {
	ctx            context.Context
	container      models.ContainerDescription
	taskService    *TaskService
	problemService *ProblemService
	variablesRepo  *repositories.ContainerVariablesRepository
}

func (h *scriptHost) AddTask(description string) (int, error) {
	parent := h.container
	full, err := h.taskService.AddTask(h.ctx, models.TaskShort{Description: description}, &parent)
	if err != nil {
		return 0, err
	}
	return full.ID, nil
}

func (h *scriptHost) AddProblem(description string) (int, error) {
	parent := h.container
	full, err := h.problemService.AddProblem(h.ctx, models.ProblemShort{Description: description}, &parent)
	if err != nil {
		return 0, err
	}
	return full.ID, nil
}

func (h *scriptHost) GetVar(name string) (string, error) {
	vars, err := h.variablesRepo.GetVariablesByContainer(h.ctx, h.container.Type, h.container.ID)
	if err != nil {
		return "", err
	}
	for _, v := range vars {
		if v.VariableName == name {
			return v.VariableValue, nil
		}
	}
	return "", nil
}

func (h *scriptHost) SetVar(name, value string) error {
	vars, err := h.variablesRepo.GetVariablesByContainer(h.ctx, h.container.Type, h.container.ID)
	if err != nil {
		return err
	}
	for _, v := range vars {
		if v.VariableName == name {
			_, err := h.variablesRepo.UpdateVariable(h.ctx, v.ID, nil, &value)
			return err
		}
	}
	_, err = h.variablesRepo.AddVariableWithValue(h.ctx, h.container.Type, h.container.ID, name, value)
	return err
}

func validateScriptMeta(name, code string, params []models.ScriptParam) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name must not be empty")
	}
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("code must not be empty")
	}
	if err := scripting.Validate(code); err != nil {
		return fmt.Errorf("syntax error: %w", err)
	}
	return validateParams(params)
}

func validateParams(params []models.ScriptParam) error {
	seen := map[string]struct{}{}
	for _, p := range params {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			return fmt.Errorf("param name must not be empty")
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("duplicate param name: %s", name)
		}
		seen[name] = struct{}{}
		switch p.Type {
		case "string", "boolean", "number":
		default:
			return fmt.Errorf("param %s has invalid type %q", name, p.Type)
		}
	}
	return nil
}

func mergeAndValidateParams(declared []models.ScriptParam, provided map[string]any) (map[string]any, error) {
	merged := map[string]any{}
	if provided == nil {
		provided = map[string]any{}
	}
	for _, p := range declared {
		val, ok := provided[p.Name]
		if !ok {
			if p.Default != nil {
				val = p.Default
			} else {
				return nil, fmt.Errorf("missing required param: %s", p.Name)
			}
		}
		coerced, err := coerceParam(p, val)
		if err != nil {
			return nil, err
		}
		merged[p.Name] = coerced
	}
	// Allow extra params not declared (pass-through) for flexibility.
	for k, v := range provided {
		if _, exists := merged[k]; !exists {
			merged[k] = v
		}
	}
	return merged, nil
}

func coerceParam(p models.ScriptParam, val any) (any, error) {
	switch p.Type {
	case "string":
		switch v := val.(type) {
		case string:
			return v, nil
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64), nil
		case bool:
			return strconv.FormatBool(v), nil
		default:
			return fmt.Sprint(v), nil
		}
	case "boolean":
		switch v := val.(type) {
		case bool:
			return v, nil
		case string:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, fmt.Errorf("param %s must be boolean", p.Name)
			}
			return b, nil
		default:
			return nil, fmt.Errorf("param %s must be boolean", p.Name)
		}
	case "number":
		switch v := val.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case string:
			n, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("param %s must be number", p.Name)
			}
			return n, nil
		default:
			return nil, fmt.Errorf("param %s must be number", p.Name)
		}
	default:
		return nil, fmt.Errorf("param %s has invalid type", p.Name)
	}
}

func toScriptFull(script *ent.Script) *models.ScriptFull {
	if script == nil {
		return nil
	}
	isGlobal := script.ContainerID == nil
	return &models.ScriptFull{
		ID:            script.ID,
		Name:          script.Name,
		Code:          script.Code,
		Description:   script.Description,
		Params:        models.FromSchemaParams(script.Params),
		IsGlobal:      isGlobal,
		ContainerType: script.ContainerType,
		ContainerID:   script.ContainerID,
		CreatedAt:     script.CreatedAt,
		UpdatedAt:     script.UpdatedAt,
	}
}

func toScriptListItem(script *ent.Script) models.ScriptListItem {
	return models.ScriptListItem{
		ID:            script.ID,
		Name:          script.Name,
		Description:   script.Description,
		IsGlobal:      script.ContainerID == nil,
		ContainerType: script.ContainerType,
		ContainerID:   script.ContainerID,
	}
}

func emptyCreated() models.ScriptRunCreated {
	return models.ScriptRunCreated{Tasks: []int{}, Problems: []int{}}
}
