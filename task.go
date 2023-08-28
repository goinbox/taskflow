package taskflow

import (
	"time"

	"github.com/goinbox/pcontext"
)

const (
	StepCodeFailure = "FAILURE"

	StepCodeSuccess = "SUCCESS"

	StepCodeJump1 = "JUMP1"
	StepCodeJump2 = "JUMP2"
	StepCodeJump3 = "JUMP3"
)

type StepFunc[T pcontext.Context] func(ctx T) (code string, err error)

type StepFailedFunc func(stepKey string, err error)

type StepConfig[T pcontext.Context] struct {
	RetryCnt   int
	RetryDelay time.Duration

	StepFunc       StepFunc[T]
	StepFailedFunc StepFailedFunc
	RouteMap       map[string]string
}

type Task[T pcontext.Context] interface {
	Name() string

	Init(in, out interface{}) error

	StepConfigMap() map[string]*StepConfig[T]
	FirstStepKey() string

	BeforeStep(stepKey string)
	AfterStep(stepKey string)

	Error() error
}
