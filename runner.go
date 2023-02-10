package taskflow

import (
	"fmt"
	"time"

	"github.com/goinbox/golog"
)

type Runner struct {
	logger golog.Logger
}

func NewRunner(logger golog.Logger) *Runner {
	r := &Runner{
		logger: logger,
	}

	return r
}

func (r *Runner) RunTask(task Task, in, out interface{}) error {
	r.logger.Notice("start runTask")
	defer func() {
		r.logger.Notice("end runTask")
	}()

	err := r.initTask(task, in, out)
	if err != nil {
		return fmt.Errorf("initTask error: %w", err)
	}

	stepConfigMap := task.StepConfigMap()
	if len(stepConfigMap) == 0 {
		r.logger.Warning("stepConfigMap's len is 0")
		return nil
	}

	nextStepKey := task.FirstStepKey()
	nextStepConfig, ok := stepConfigMap[nextStepKey]
	if !ok {
		r.logger.Error("firstStep not exists", &golog.Field{
			Key:   LogFieldKeyStepKey,
			Value: nextStepKey,
		})
		return nil
	}

	for {
		task.BeforeStep(nextStepKey)
		curStepKey := nextStepKey
		nextStepKey = r.runStep(nextStepKey, nextStepConfig)
		task.AfterStep(curStepKey)

		if nextStepKey == "" {
			break
		}
		nextStepConfig = stepConfigMap[nextStepKey]
		if nextStepConfig == nil {
			break
		}
	}

	return nil
}

func (r *Runner) initTask(task Task, in, out interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("recover from %s", fmt.Sprint(e))
		}
	}()

	return task.Init(in, out)
}

func (r *Runner) runStep(stepKey string, config *StepConfig) (nextStepKey string) {
	stepFunc := config.StepFunc
	logger := r.logger.With(&golog.Field{
		Key:   LogFieldKeyStepKey,
		Value: stepKey,
	})

	logger.Notice("start runStep")

	code, err := r.runStepFunc(stepFunc)
	if err != nil {
		logger.Error("runStep error", golog.ErrorField(err))
		if code == "" {
			if config.RetryCnt > 0 {
				code, err = r.retryStep(logger, config, stepFunc)
			} else {
				code = StepCodeFailure
			}
		}
	}

	if code == StepCodeFailure {
		if config.StepFailedFunc != nil {
			logger.Notice("run StepFailedFunc")
			config.StepFailedFunc(stepKey, err)
		}
	}

	nextStepKey = config.RouteMap[code]

	logger.Notice("end runStep", []*golog.Field{
		{
			Key:   "code",
			Value: code,
		},
		{
			Key:   "nextStepKey",
			Value: nextStepKey,
		},
	}...)

	return nextStepKey
}

func (r *Runner) runStepFunc(f StepFunc) (code string, err error) {
	defer func() {
		if e := recover(); e != nil {
			code = StepCodeFailure
			err = fmt.Errorf("recover from %s", fmt.Sprint(e))
		}
	}()

	return f()
}

func (r *Runner) retryStep(logger golog.Logger, config *StepConfig, stepFunc StepFunc) (code string, err error) {
	for i := 0; i < config.RetryCnt; i++ {
		logger.Notice("wait retry runStep")

		time.Sleep(config.RetryDelay)

		logger.Notice("retry runStep", []*golog.Field{
			{
				Key:   "RetryNo",
				Value: i + 1,
			},
			{
				Key:   "RetryCount",
				Value: config.RetryCnt,
			},
		}...)

		code, err = r.runStepFunc(stepFunc)
		if err == nil {
			return code, nil
		}

		logger.Error("runStep error", golog.ErrorField(err))
		if code != "" {
			return code, err
		}

	}
	if err != nil {
		code = StepCodeFailure
	}

	return code, err
}

func (r *Runner) TaskGraph(task Task, codes ...string) string {
	filterCode := false
	codeMap := make(map[string]bool)
	if len(codes) > 0 {
		filterCode = true
		for _, code := range codes {
			codeMap[code] = true
		}
	}

	result := "```mermaid\nflowchart TD\n"
	for curStep, config := range task.StepConfigMap() {
		for code, nextStep := range config.RouteMap {
			if filterCode {
				if _, ok := codeMap[code]; !ok {
					continue
				}
			}
			if nextStep == "" {
				nextStep = "finish"
			}
			result += fmt.Sprintf("%s --%s--> %s\n", curStep, code, nextStep)
		}
	}

	result += fmt.Sprintf("style %s fill:#f9f\n", "finish")
	result += "```"

	return result
}
