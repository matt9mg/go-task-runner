# Go Task Runner

Runs go routines or a group of go routines and returns its outcomes

### Installation

```
go get github.com/matt9mg/go-task-runner
```

### Examples

Running a single task

```go
handler := runner.NewRunner()

err := handler.AddTask(func () (any, error) {
return "here", nil
}, "test")

if err != nil {
log.Fatalln(err)
}

responses := handler.Run()

log.Print(*responses["test"].Value) // "here"
log.Print(*responses["test"].Error) // nil or error string
```

Running multiple tasks of different types

```go
handler := runner.NewRunner()

err := handler.AddTask(func () (any, error) {
    return "here", nil
}, "test")

if err != nil {
log.Fatalln(err)
}

err = handler.AddTaskWithTimeout(func () (any, error) {
    time.Sleep(time.Second * 2)
    return 1, nil
}, "test2", context.TODO(), time.Second)

if err != nil {
    log.Fatalln(err)
}

responses := handler.Run()

log.Print(*responses["test"].Value) // "here"
log.Print(*responses["test"].Error) // nil or error string
log.Print(*responses["test2"].Value) // nil
log.Print(*responses["test2"].Error.Error()) // context deadline exceeded
```

With structs
```go
handler := runner.NewRunner()

err := handler.AddTask(func () (any, error) {
    return return struct {
A int
}{A: 1}, nil
}, "test")

if err != nil {
log.Fatalln(err)
}

log.Println(*response["test"].Value) // struct{A: 1}
log.Print(*responses["test"].Error) // nil or error string
```