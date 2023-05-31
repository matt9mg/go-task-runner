package runner_test

import (
	"context"
	"github.com/matt9mg/go-task-runner"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRunner_AddTask(t *testing.T) {
	_assert := assert.New(t)

	handler := runner.NewRunner()

	err := handler.AddTask(func() (any, error) {
		return "here", nil
	}, "test")

	_assert.Nil(err)

	responses := handler.Run()

	_assert.Len(responses, 1)
	_assert.Equal("here", responses["test"].Value)
	_assert.Nil(responses["test"].Error)

}

func TestRunner_AddMultipleTasks(t *testing.T) {
	_assert := assert.New(t)

	handler := runner.NewRunner()

	err := handler.AddTask(func() (any, error) {
		return "here", nil
	}, "test")

	_assert.Nil(err)

	err = handler.AddTask(func() (any, error) {
		return 1, nil
	}, "test2")

	_assert.Nil(err)

	responses := handler.Run()

	_assert.Len(responses, 2)
	_assert.Equal("here", responses["test"].Value)
	_assert.Nil(responses["test"].Error)
	_assert.Equal(1, responses["test2"].Value)
	_assert.Nil(responses["test2"].Error)

}

func TestRunner_AddMultipleTasks_Dupes(t *testing.T) {
	_assert := assert.New(t)

	handler := runner.NewRunner()

	err := handler.AddTask(func() (any, error) {
		return "here", nil
	}, "test")

	_assert.Nil(err)

	err = handler.AddTask(func() (any, error) {
		return 1, nil
	}, "test")

	_assert.NotNil(err)
	_assert.Equal("task with name test already added to task queue", err.Error())
}

func TestRunner_AddMultipleTasksWithTimeout(t *testing.T) {
	_assert := assert.New(t)

	handler := runner.NewRunner()

	err := handler.AddTask(func() (any, error) {
		return "here", nil
	}, "test")

	_assert.Nil(err)

	err = handler.AddTask(func() (any, error) {
		return "here2", nil
	}, "test2")

	_assert.Nil(err)

	err = handler.AddTaskWithTimeout(func() (any, error) {
		return 1, nil
	}, "test3", context.TODO(), time.Second)

	_assert.Nil(err)

	err = handler.AddTaskWithTimeout(func() (any, error) {
		return struct {
			A int
		}{A: 1}, nil
	}, "test4", context.TODO(), time.Second)

	_assert.Nil(err)

	responses := handler.Run()

	_assert.Len(responses, 4)
	_assert.Equal("here", responses["test"].Value)
	_assert.Nil(responses["test"].Error)
	_assert.Equal("here2", responses["test2"].Value)
	_assert.Nil(responses["test2"].Error)
	_assert.Equal(1, responses["test3"].Value)
	_assert.Nil(responses["test3"].Error)
	_assert.Equal(1, responses["test4"].Value.(struct{ A int }).A)
	_assert.Nil(responses["test4"].Error)
}

func TestRunner_AddMultipleTasksWithTimeoutDupes(t *testing.T) {
	_assert := assert.New(t)

	handler := runner.NewRunner()

	err := handler.AddTask(func() (any, error) {
		return "here", nil
	}, "test")

	_assert.Nil(err)

	err = handler.AddTaskWithTimeout(func() (any, error) {
		return 1, nil
	}, "test", context.TODO(), time.Second)

	_assert.Equal("task with name test already added to task queue", err.Error())
}

func TestRunner_AddMultipleTasksWithTimeoutTimeout(t *testing.T) {
	_assert := assert.New(t)

	handler := runner.NewRunner()

	err := handler.AddTask(func() (any, error) {
		return "here", nil
	}, "test")

	_assert.Nil(err)

	err = handler.AddTaskWithTimeout(func() (any, error) {
		time.Sleep(time.Second * 2)
		return 1, nil
	}, "test2", context.TODO(), time.Second)

	_assert.Nil(err)

	responses := handler.Run()

	_assert.Len(responses, 2)
	_assert.Equal("here", responses["test"].Value)
	_assert.Nil(responses["test"].Error)

	_assert.Equal(nil, responses["test2"].Value)
	_assert.Equal("context deadline exceeded", responses["test2"].Error.Error())
}

func TestRunner_CancelAll(t *testing.T) {
	_assert := assert.New(t)

	handler := runner.NewRunner()

	err := handler.AddTask(func() (any, error) {
		return "here", nil
	}, "test")

	_assert.Nil(err)

	err = handler.AddTaskWithTimeout(func() (any, error) {
		time.Sleep(time.Second * 2)
		return 1, nil
	}, "test2", context.TODO(), time.Second)

	_assert.Nil(err)

	handler.CancelAll()
	responses := handler.Run()

	_assert.Len(responses, 2)
	_assert.Nil(responses["test"].Value)
	_assert.Equal("context canceled", responses["test"].Error.Error())

	_assert.Nil(responses["test2"].Value)
	_assert.Equal("context canceled", responses["test2"].Error.Error())
}

func TestRunner_CancelByName(t *testing.T) {
	_assert := assert.New(t)

	handler := runner.NewRunner()

	err := handler.AddTask(func() (any, error) {
		return "here", nil
	}, "test")

	_assert.Nil(err)

	err = handler.AddTaskWithTimeout(func() (any, error) {
		return 1, nil
	}, "test2", context.TODO(), time.Second)

	_assert.Nil(err)

	handler.CancelTaskByName("test")
	responses := handler.Run()

	_assert.Len(responses, 2)
	_assert.Nil(responses["test"].Value)
	_assert.Equal("context canceled", responses["test"].Error.Error())

	_assert.Equal(1, responses["test2"].Value)
	_assert.Nil(responses["test2"].Error)
}
