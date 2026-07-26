package scripting

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
)

const defaultTimeout = 5 * time.Second

// HostAPI is implemented by the scripts service to expose host actions to JS.
type HostAPI interface {
	AddTask(description string) (int, error)
	AddProblem(description string) (int, error)
	GetVar(name string) (string, error)
	SetVar(name, value string) error
}

// RunResult is returned after a successful script execution.
type RunResult struct {
	CreatedTasks    []int
	CreatedProblems []int
}

type hostBridge struct {
	mu      sync.Mutex
	api     HostAPI
	tasks   []int
	problems []int
	err     error
}

func (h *hostBridge) setErr(err error) {
	if err == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.err == nil {
		h.err = err
	}
}

func (h *hostBridge) getErr() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.err
}

// Validate parses script code without executing host APIs.
func Validate(code string) error {
	_, err := goja.Compile("script", code, false)
	return err
}

// Run executes script code with the given params and host API.
func Run(ctx context.Context, code string, params map[string]any, api HostAPI) (*RunResult, error) {
	if err := Validate(code); err != nil {
		return nil, fmt.Errorf("syntax error: %w", err)
	}

	vm := goja.New()
	bridge := &hostBridge{api: api}

	if params == nil {
		params = map[string]any{}
	}
	if err := vm.Set("params", params); err != nil {
		return nil, err
	}
	// Also expose each param as a top-level binding so scripts can use
	// `if (wasExecuted)` in addition to `if (params.wasExecuted)`.
	reserved := map[string]struct{}{
		"params": {}, "container": {}, "vars": {},
	}
	for name, value := range params {
		if name == "" {
			continue
		}
		if _, taken := reserved[name]; taken {
			continue
		}
		if err := vm.Set(name, value); err != nil {
			return nil, err
		}
	}

	containerObj := vm.NewObject()
	_ = containerObj.Set("addTask", func(call goja.FunctionCall) goja.Value {
		if bridge.getErr() != nil {
			panic(vm.ToValue(bridge.getErr().Error()))
		}
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("container.addTask requires a description"))
		}
		desc := call.Argument(0).String()
		id, err := api.AddTask(desc)
		if err != nil {
			bridge.setErr(err)
			panic(vm.ToValue(err.Error()))
		}
		bridge.mu.Lock()
		bridge.tasks = append(bridge.tasks, id)
		bridge.mu.Unlock()
		return vm.ToValue(id)
	})
	_ = containerObj.Set("addProblem", func(call goja.FunctionCall) goja.Value {
		if bridge.getErr() != nil {
			panic(vm.ToValue(bridge.getErr().Error()))
		}
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("container.addProblem requires a description"))
		}
		desc := call.Argument(0).String()
		id, err := api.AddProblem(desc)
		if err != nil {
			bridge.setErr(err)
			panic(vm.ToValue(err.Error()))
		}
		bridge.mu.Lock()
		bridge.problems = append(bridge.problems, id)
		bridge.mu.Unlock()
		return vm.ToValue(id)
	})
	if err := vm.Set("container", containerObj); err != nil {
		return nil, err
	}

	varsObj := vm.NewObject()
	_ = varsObj.Set("get", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("vars.get requires a name"))
		}
		name := call.Argument(0).String()
		val, err := api.GetVar(name)
		if err != nil {
			bridge.setErr(err)
			panic(vm.ToValue(err.Error()))
		}
		return vm.ToValue(val)
	})
	_ = varsObj.Set("set", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.ToValue("vars.set requires name and value"))
		}
		name := call.Argument(0).String()
		value := call.Argument(1).String()
		if err := api.SetVar(name, value); err != nil {
			bridge.setErr(err)
			panic(vm.ToValue(err.Error()))
		}
		return goja.Undefined()
	})
	if err := vm.Set("vars", varsObj); err != nil {
		return nil, err
	}

	runCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("%v", r)
			}
		}()
		_, err := vm.RunString(code)
		done <- err
	}()

	select {
	case <-runCtx.Done():
		vm.Interrupt("script execution timed out")
		<-done
		return nil, fmt.Errorf("script execution timed out")
	case err := <-done:
		if err != nil {
			if hostErr := bridge.getErr(); hostErr != nil {
				return nil, hostErr
			}
			return nil, err
		}
	}

	if hostErr := bridge.getErr(); hostErr != nil {
		return nil, hostErr
	}

	bridge.mu.Lock()
	defer bridge.mu.Unlock()
	return &RunResult{
		CreatedTasks:    append([]int(nil), bridge.tasks...),
		CreatedProblems: append([]int(nil), bridge.problems...),
	}, nil
}
