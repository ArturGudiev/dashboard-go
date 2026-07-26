package scripting

import (
	"context"
	"testing"
)

type mockHost struct {
	tasks    []string
	problems []string
	vars     map[string]string
}

func (m *mockHost) AddTask(description string) (int, error) {
	m.tasks = append(m.tasks, description)
	return len(m.tasks), nil
}

func (m *mockHost) AddProblem(description string) (int, error) {
	m.problems = append(m.problems, description)
	return len(m.problems), nil
}

func (m *mockHost) GetVar(name string) (string, error) {
	return m.vars[name], nil
}

func (m *mockHost) SetVar(name, value string) error {
	if m.vars == nil {
		m.vars = map[string]string{}
	}
	m.vars[name] = value
	return nil
}

func TestValidateOK(t *testing.T) {
	if err := Validate(`container.addTask("x");`); err != nil {
		t.Fatalf("expected valid syntax, got %v", err)
	}
}

func TestValidateBad(t *testing.T) {
	if err := Validate(`if (`); err == nil {
		t.Fatal("expected syntax error")
	}
}

func TestRunAddTaskAndParams(t *testing.T) {
	host := &mockHost{}
	code := `
container.addTask("base");
if (params.myVar === true) {
  container.addTask("case 1 task");
} else {
  container.addTask("case 2 task");
}
container.addProblem("p1");
vars.set("flag", "1");
`
	result, err := Run(context.Background(), code, map[string]any{"myVar": true}, host)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(result.CreatedTasks) != 2 || len(result.CreatedProblems) != 1 {
		t.Fatalf("unexpected created: %+v", result)
	}
	if host.vars["flag"] != "1" {
		t.Fatalf("expected var set, got %v", host.vars)
	}
	if host.tasks[1] != "case 1 task" {
		t.Fatalf("expected case 1, got %q", host.tasks[1])
	}
}

func TestRunBareParamBinding(t *testing.T) {
	host := &mockHost{}
	code := `
if (wasExecuted) {
  container.addTask("wasExecuted was true");
} else {
  container.addTask("wasExecuted was false");
}
`
	result, err := Run(context.Background(), code, map[string]any{"wasExecuted": true}, host)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(result.CreatedTasks) != 1 || host.tasks[0] != "wasExecuted was true" {
		t.Fatalf("unexpected created: %+v tasks=%v", result, host.tasks)
	}
}
