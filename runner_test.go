package taskflow

import (
	"fmt"
	"testing"

	"github.com/goinbox/golog"
)

func runTaskUseP(in *demoTaskIn) *demoTaskOut {
	w, _ := golog.NewFileWriter("/dev/stdout", 0)
	logger := golog.NewSimpleLogger(w, golog.NewSimpleFormater())

	out := new(demoTaskOut)
	err := NewRunner(logger).RunTask(new(demoTask), in, out)

	fmt.Println("error", err)

	return out
}

func TestRunTaskUseP1(t *testing.T) {
	out := runTaskUseP(&demoTaskIn{
		id:          1,
		failureStep: "first",
	})

	t.Log("out", out)
}

func TestRunTaskUseP2(t *testing.T) {
	out := runTaskUseP(&demoTaskIn{
		id:          2,
		failureStep: "",
	})

	t.Log("out", out)
}

func TestRunTaskUseP3(t *testing.T) {
	out := runTaskUseP(&demoTaskIn{
		id:          3,
		failureStep: "second",
	})

	t.Log("out", out)
}

func TestTaskGraph(t *testing.T) {
	graph := NewRunner(nil).TaskGraph(new(demoTask))

	t.Log(graph)
}
