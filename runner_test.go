package taskflow

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/goinbox/golog"
)

func TestTaskGraph(t *testing.T) {
	graph := NewRunner(nil).TaskGraph(new(demoTask))

	t.Log(graph)
}

func runTaskUseP(in *demoTaskIn) (*demoTaskOut, *Runner) {
	w, _ := golog.NewFileWriter("/dev/stdout", 0)
	logger := golog.NewSimpleLogger(w, golog.NewSimpleFormater())

	out := new(demoTaskOut)
	runner := NewRunner(logger)
	err := runner.RunTask(new(demoTask), in, out)

	fmt.Println("error", err)

	return out, runner
}

func TestRunTaskUseP1(t *testing.T) {
	out, _ := runTaskUseP(&demoTaskIn{
		id:          1,
		failureStep: "first",
	})

	t.Log("out", out)
}

func TestRunTaskUseP2(t *testing.T) {
	out, _ := runTaskUseP(&demoTaskIn{
		id:          2,
		failureStep: "",
	})

	t.Log("out", out)
}

func TestRunTaskUseP3(t *testing.T) {
	out, _ := runTaskUseP(&demoTaskIn{
		id:          3,
		failureStep: "second",
	})

	t.Log("out", out)
}

func TestTaskRunSteps(t *testing.T) {
	_, runner := runTaskUseP(&demoTaskIn{
		id:          1,
		failureStep: "first",
	})

	for i, runStep := range runner.runSteps {
		t.Log(i, runStep)
	}

	content, _ := json.Marshal(runner.TaskRunSteps())
	t.Log(string(content))

	w, _ := golog.NewFileWriter("/dev/stdout", 0)
	logger := golog.NewSimpleLogger(w, golog.NewJsonFormater())
	logger.Info("log run steps", &golog.Field{
		Key:   "RunSteps",
		Value: string(content),
	})
}

func TestTaskGraphRunSteps(t *testing.T) {
	runSteps := []*RunStep{
		{
			StepKey:  "first",
			StepCode: "SUCCESS",
		},
		{
			StepKey:  "second",
			StepCode: "JUMP2",
		},
		{
			StepKey:  "jump",
			StepCode: "SUCCESS",
		},
	}

	t.Log(NewRunner(nil).TaskGraphRunSteps(new(demoTask), runSteps))
}

func TestTaskGraphRunStepsFromJson(t *testing.T) {
	s := `
[
    {
        "StepKey": "first",
        "StepCode": "JUMP1"
    },
    {
        "StepKey": "jump",
        "StepCode": "SUCCESS"
    }
]
`
	graph, err := NewRunner(nil).TaskGraphRunStepsFromJson(new(demoTask), []byte(s))
	t.Log(graph, err)
}
