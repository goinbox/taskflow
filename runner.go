package taskflow

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/goinbox/golog"
)

type GraphConfig struct {
	FinishStepKey string

	StartStyleColor   string
	FinishStyleColor  string
	RunStepStyleColor string
}

type RunStep struct {
	StepKey  string
	StepCode string
}

type Runner struct {
	GraphConfig *GraphConfig

	logger   golog.Logger
	runSteps []*RunStep
}

func NewRunner(logger golog.Logger) *Runner {
	r := &Runner{
		logger: logger,
		GraphConfig: &GraphConfig{
			FinishStepKey:     "finish",
			StartStyleColor:   "#b57edc",
			FinishStyleColor:  "#74c365",
			RunStepStyleColor: "#ff9966",
		},
	}

	return r
}

func (r *Runner) RunTask(task Task, in, out interface{}) error {
	r.logger.Notice("start runTask")
	defer func() {
		r.logger.Notice("end runTask")

		r.logger.Info("task run steps", &golog.Field{
			Key:   "RunSteps",
			Value: r.runSteps,
		})
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
			err = fmt.Errorf("recover from %v, stack: %s", e, string(debug.Stack()))
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
	r.runSteps = append(r.runSteps, &RunStep{
		StepKey:  stepKey,
		StepCode: code,
	})

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
			err = fmt.Errorf("recover from %v, stack: %s", e, string(debug.Stack()))
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

type graphRowFunc func(curStep, code, nextStep string) string

func (r *Runner) graphContent(task Task, showCodes []string, grf graphRowFunc) string {
	var graph string

	filterCode := false
	showCodeMap := make(map[string]bool)
	if len(showCodes) > 0 {
		filterCode = true
		for _, code := range showCodes {
			showCodeMap[code] = true
		}
	}
	for curStep, config := range task.StepConfigMap() {
		for code, nextStep := range config.RouteMap {
			if filterCode {
				if _, ok := showCodeMap[code]; !ok {
					continue
				}
			}
			if nextStep == "" {
				nextStep = r.GraphConfig.FinishStepKey
			}
			graph += grf(curStep, code, nextStep)
		}
	}

	return graph
}

func (r *Runner) drawGraph(content, style string) string {
	style += fmt.Sprintf("style %s fill:%s\n", r.GraphConfig.FinishStepKey, r.GraphConfig.FinishStyleColor)

	graph := "```mermaid\nflowchart TD\n"
	graph += content + style
	graph += "```"

	return graph
}

func (r *Runner) TaskGraph(task Task, showCodes ...string) string {
	var style string
	style += fmt.Sprintf("style %s fill:%s\n", task.FirstStepKey(), r.GraphConfig.StartStyleColor)

	grf := func(curStep, code, nextStep string) string {
		return fmt.Sprintf("%s --%s--> %s\n", curStep, code, nextStep)
	}

	return r.drawGraph(r.graphContent(task, showCodes, grf), style)
}

func (r *Runner) TaskRunSteps() []*RunStep {
	return r.runSteps
}

func (r *Runner) TaskGraphRunSteps(task Task, runSteps []*RunStep, showCodes ...string) string {
	runStepMap := make(map[string]bool)
	var style string
	for _, runStep := range runSteps {
		runStepMap[runStep.StepKey+runStep.StepCode] = true
		style += fmt.Sprintf("style %s fill:%s\n", runStep.StepKey, r.GraphConfig.RunStepStyleColor)
	}

	grf := func(curStep, code, nextStep string) string {
		_, ok := runStepMap[curStep+code]
		if ok {
			return fmt.Sprintf("%s ==%s==> %s\n", curStep, code, nextStep)
		}
		return fmt.Sprintf("%s -.%s.-> %s\n", curStep, code, nextStep)
	}

	return r.drawGraph(r.graphContent(task, showCodes, grf), style)
}

func (r *Runner) TaskGraphRunStepsFromJson(task Task, runStepsJson []byte, showCodes ...string) (string, error) {
	var runSteps []*RunStep
	err := json.Unmarshal(runStepsJson, &runSteps)
	if err != nil {
		return "", fmt.Errorf("json.Unmarshal error: %w", err)
	}

	return r.TaskGraphRunSteps(task, runSteps, showCodes...), nil
}
