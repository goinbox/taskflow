package taskflow

import (
	"fmt"
	"time"
)

type demoTaskIn struct {
	id          int
	failureStep string
}

type demoTaskOut struct {
	finalStep string
}

type demoTask struct {
	in  *demoTaskIn
	out *demoTaskOut

	data struct {
		cnt int
	}
}

func (t *demoTask) Name() string {
	return "demo"
}

func (t *demoTask) Init(in, out interface{}) error {
	t.in = in.(*demoTaskIn)
	t.out = out.(*demoTaskOut)

	return nil
}

func (t *demoTask) StepConfigMap() map[string]*StepConfig {
	return map[string]*StepConfig{
		"first": {
			RetryCnt:       1,
			RetryDelay:     time.Second * 1,
			StepFunc:       t.firstStep,
			StepFailedFunc: t.stepFailedFunc,
			RouteMap: map[string]string{
				StepCodeSuccess: "second",
				StepCodeFailure: "failure",
				StepCodeJump1:   "jump",
			},
		},
		"second": {
			RetryCnt:   2,
			RetryDelay: time.Second * 2,
			StepFunc:   t.secondStep,
			RouteMap: map[string]string{
				StepCodeSuccess: "",
				StepCodeFailure: "failure",
				StepCodeJump2:   "jump",
			},
		},
		"failure": {
			RetryCnt:   0,
			RetryDelay: 0,
			StepFunc:   t.failureStep,
			RouteMap: map[string]string{
				StepCodeSuccess: "",
				StepCodeJump3:   "jump",
			},
		},
		"jump": {
			RetryCnt:   0,
			RetryDelay: 0,
			StepFunc:   t.jumpStep,
			RouteMap: map[string]string{
				StepCodeSuccess: "",
			},
		},
	}
}

func (t *demoTask) FirstStepKey() string {
	return "first"
}

func (t *demoTask) BeforeStep(stepKey string) {
	fmt.Println("BeforeStep", stepKey)

	t.out.finalStep = stepKey
}

func (t *demoTask) AfterStep(stepKey string) {
	fmt.Println("AfterStep", stepKey)
}

func (t *demoTask) Error() error {
	return nil
}

func (t *demoTask) firstStep() (string, error) {
	fmt.Println("in firstStep")

	if t.data.cnt == 0 {
		t.data.cnt++
		return "", fmt.Errorf("firstStep error")
	}

	if t.in.id == 1 {
		return StepCodeJump1, nil
	}

	if t.in.failureStep == "first" {
		return "", fmt.Errorf("failure in firstStep")
	}

	return StepCodeSuccess, nil
}

func (t *demoTask) secondStep() (string, error) {
	fmt.Println("in secondStep")

	defer func() {
		t.data.cnt++
	}()

	if t.data.cnt < 2 {
		panic("panic in secondStep")
	}

	if t.in.id == 2 {
		return StepCodeJump2, nil
	}

	if t.in.failureStep == "second" {
		return "", fmt.Errorf("failure in secondStep")
	}

	return StepCodeSuccess, nil
}

func (t *demoTask) failureStep() (string, error) {
	fmt.Println("in failureStep")

	if t.in.failureStep == "jump" {
		return StepCodeJump3, nil
	}

	return StepCodeSuccess, nil
}

func (t *demoTask) jumpStep() (string, error) {
	fmt.Println("in jumpStep")

	fmt.Println("final cnt", t.data.cnt)

	return StepCodeSuccess, nil
}

func (t *demoTask) stepFailedFunc(stepKey string, err error) {
	fmt.Println("step failed error:", stepKey, err)
}
