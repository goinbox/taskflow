package taskflow

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/goinbox/golog"
	"github.com/goinbox/pcontext"

	"go.opentelemetry.io/otel/trace"
)

func startTrace(ctx pcontext.Context, spanName string, opts ...trace.SpanStartOption) (pcontext.Context, trace.Span) {
	fmt.Println("---------- start trace", spanName)
	_, span := trace.NewNoopTracerProvider().Tracer("").Start(ctx, spanName, opts...)
	return ctx, span
}

func TestTaskGraph(t *testing.T) {
	graph := NewRunner[pcontext.Context]().TaskGraph(new(demoTask))

	t.Log(graph)
}

func runTaskUseP(in *demoTaskIn) (*demoTaskOut, *Runner[pcontext.Context]) {
	w, _ := golog.NewFileWriter("/dev/stdout", 0)
	logger := golog.NewSimpleLogger(w, golog.NewSimpleFormater())
	ctx := pcontext.NewSimpleContext(logger)

	out := new(demoTaskOut)
	runner := NewRunner[pcontext.Context]().SetStartTraceFunc(startTrace)
	err := runner.RunTask(ctx, new(demoTask), in, out)

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

	t.Log(NewRunner[pcontext.Context]().TaskGraphRunSteps(new(demoTask), runSteps))
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
	graph, err := NewRunner[pcontext.Context]().TaskGraphRunStepsFromJson(new(demoTask), []byte(s))
	t.Log(graph, err)
}
