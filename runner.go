package runner

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Runner struct {
	tasks    map[string]*RunnerTask
	reponses []any
}

type RunnerTask struct {
	task       RunnerFunc
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type response struct {
	key   string
	Value any
	Error error
}

type RunnerFunc func() (any, error)

func NewRunner() *Runner {
	return &Runner{
		tasks: make(map[string]*RunnerTask),
	}
}

// AddTask specify the function routine and the task name
// if the task already exists with the same name an error is returned
func (r *Runner) AddTask(rf RunnerFunc, taskName string) error {
	if _, ok := r.tasks[taskName]; ok {
		return fmt.Errorf("task with name %s already added to task queue", taskName)
	}

	ctx, cancel := context.WithCancel(context.Background())

	r.tasks[taskName] = &RunnerTask{
		task:       rf,
		ctx:        ctx,
		cancelFunc: cancel,
	}

	return nil
}

// CancelTaskByName takes the task name and cancels the underlying context or to prevent a context leak
func (r *Runner) CancelTaskByName(taskName string) {
	if _, ok := r.tasks[taskName]; ok {
		r.tasks[taskName].cancelFunc()
	}
}

// CancelAll cancels all tasks for their underlying context or to prevent a context leak
func (r *Runner) CancelAll() {
	for _, t := range r.tasks {
		t.cancelFunc()
	}
}

// AddTaskWithTimeout specify the function routine, task name, parent context and a time duration
// time duration is how long it should run before it's automatically cancelled
func (r *Runner) AddTaskWithTimeout(rf RunnerFunc, taskName string, ctx context.Context, d time.Duration) error {
	if _, ok := r.tasks[taskName]; ok {
		return fmt.Errorf("task with name %s already added to task queue", taskName)
	}

	subCtx, cancel := context.WithTimeout(ctx, d)

	r.tasks[taskName] = &RunnerTask{
		task:       rf,
		ctx:        subCtx,
		cancelFunc: cancel,
	}

	return nil
}

// Run runs all the tasks as routines and returns a mapped response where the key is the task name
func (r *Runner) Run() map[string]*response {
	taskLen := len(r.tasks)
	responses := make(map[string]*response, taskLen)
	respCh := make(chan *response, taskLen)
	exitChan := make(chan struct{}, 1)

	defer close(respCh)

	wg := &sync.WaitGroup{}
	wg.Add(taskLen)

	go func(respCh chan<- *response, exitChan chan<- struct{}, wg *sync.WaitGroup) {
		for runnerId, rt := range r.tasks {
			go r.spawn(rt, runnerId, respCh, wg)
		}

		wg.Wait()
		close(exitChan)
	}(respCh, exitChan, wg)

complete:
	for {
		select {
		case r := <-respCh:
			responses[r.key] = r
			break
		case <-exitChan:
			break complete
		}
	}

	r.tasks = make(map[string]*RunnerTask)

	return responses
}

func (r *Runner) spawn(rt *RunnerTask, runnerId string, respCh chan<- *response, wg *sync.WaitGroup) {
	defer wg.Done()
	subChan := make(chan *response, 1)
	cancel := make(chan struct{}, 1)

	if err := rt.ctx.Err(); err != context.Canceled {
		go func(runnerId string, subChan chan<- *response, cancel chan struct{}) {
			val, err := rt.task()

			subChan <- &response{
				key:   runnerId,
				Value: val,
				Error: err,
			}
		}(runnerId, subChan, cancel)
	}

	select {
	case <-rt.ctx.Done():
		respCh <- &response{
			key:   runnerId,
			Value: nil,
			Error: rt.ctx.Err(),
		}
		break
	case subResp := <-subChan:
		respCh <- subResp
		break
	}

	close(subChan)
}
