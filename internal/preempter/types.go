package preempter

import (
	"context"
	"github.com/ecodeclub/ecron/internal/task"
)

//go:generate mockgen -source=./types.go -package=preemptermocks -destination=./mocks/preempter.mock.go
type Preempter interface {
	Preempt(ctx context.Context) (task.Task, CancelFn, error)
}

// CancelFn 停止续约，释放任务以及计算任务下一次执行时间
type CancelFn func()
