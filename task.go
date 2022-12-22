package taskflow

import (
	"time"
)

const (
	StepCodeFailure = "FAILURE"

	StepCodeSuccess = "SUCCESS"

	StepCodeJump1 = "JUMP1"
	StepCodeJump2 = "JUMP2"
	StepCodeJump3 = "JUMP3"

	LogFieldKeyStepKey = "StepKey"
)

type StepFunc func() (string, error)

type StepFailedFunc func(stepKey string, err error)

type StepConfig struct {
	RetryCnt   int
	RetryDelay time.Duration

	StepFunc       StepFunc
	StepFailedFunc StepFailedFunc
	RouteMap       map[string]string
}

type Task interface {
	Name() string

	Init(in, out interface{}) error

	StepConfigMap() map[string]*StepConfig
	FirstStepKey() string

	BeforeStep(stepKey string)
	AfterStep(stepKey string)

	Error() error
}
