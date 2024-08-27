package scheduler

import (
	"context"
	"errors"
	"github.com/ecodeclub/ecron/internal/executor"
	"github.com/ecodeclub/ecron/internal/storage"
	"github.com/ecodeclub/ecron/internal/task"
	"golang.org/x/sync/semaphore"
	"log/slog"
	"time"
)

type PreemptScheduler struct {
	dao             storage.TaskDAO
	executionDAO    storage.ExecutionDAO
	executors       map[string]executor.Executor
	refreshInterval time.Duration
	limiter         *semaphore.Weighted
	logger          *slog.Logger
}

func NewPreemptScheduler(dao storage.TaskDAO, executionDAO storage.ExecutionDAO,
	refreshInterval time.Duration, limiter *semaphore.Weighted, logger *slog.Logger) *PreemptScheduler {
	return &PreemptScheduler{
		dao:             dao,
		executionDAO:    executionDAO,
		refreshInterval: refreshInterval,
		limiter:         limiter,
		executors:       make(map[string]executor.Executor),
		logger:          logger,
	}
}

func (p *PreemptScheduler) RegisterExecutor(execs ...executor.Executor) {
	for _, exec := range execs {
		p.executors[exec.Name()] = exec
	}
}

func (p *PreemptScheduler) Schedule(ctx context.Context) error {
	for {
		err := p.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}

		ctx2, cancel := context.WithTimeout(ctx, time.Second*3)
		t, err := p.dao.Preempt(ctx2)
		cancel()
		if err != nil {
			continue
		}
		exec, ok := p.executors[t.Executor]
		if !ok {
			p.logger.Error("找不到任务的执行器",
				slog.Int64("TaskID", t.ID),
				slog.String("Executor", t.Executor))
			continue
		}

		go p.doTask(ctx, t, exec)
	}
}

func (p *PreemptScheduler) doTask(ctx context.Context, t task.Task, exec executor.Executor) {
	defer p.limiter.Release(1)
	defer p.releaseTask(t)

	// 任务执行超时配置
	timeout := exec.TaskTimeout(t)

	eid, err := p.markStatus(t.ID, task.ExecStatusStarted)
	if err != nil {
		// 这里我直接返回，如果只是网络抖动，那么任务释放后，不修改下一次执行时间，该任务可以立刻再次被抢占执行。
		// 如果是数据库异常，则无法记录任务执行情况，那么放弃这一次执行。
		return
	}
	// 控制任务执行时长
	execCtx, execCancel := context.WithDeadline(ctx, time.Now().Add(timeout))
	defer execCancel()

	ticker := time.NewTicker(p.refreshInterval)
	defer ticker.Stop()
	go func() {
		refreshCtx, refreshCancel := context.WithTimeout(execCtx, time.Second*3)
		err := p.refreshTask(refreshCtx, ticker, t.ID)
		refreshCancel()
		if err != nil {
			// 续约失败时，通知用户停止执行任务
			execCancel()
		}
	}()

	status, _ := exec.Run(execCtx, t, eid)
	defer p.setNextTime(t)

	switch status {
	case task.ExecStatusSuccess:
		_, _ = p.markStatus(t.ID, task.ExecStatusSuccess)
	case task.ExecStatusDeadlineExceeded:
		_, _ = p.markStatus(t.ID, task.ExecStatusDeadlineExceeded)
	case task.ExecStatusCancelled:
		_, _ = p.markStatus(t.ID, task.ExecStatusCancelled)
	case task.ExecStatusRunning:
		_, _ = p.markStatus(t.ID, task.ExecStatusRunning)
		p.explore(execCtx, exec, eid, t)
	default:
		_, _ = p.markStatus(t.ID, task.ExecStatusFailed)
	}
}

func (p *PreemptScheduler) explore(ctx context.Context, exec executor.Executor, eid int64, t task.Task) {
	ch := exec.Explore(ctx, eid, t)
	if ch == nil {
		return
	}
	end := false
	for !end {
		select {
		case <-ctx.Done():
			// 主动取消或者超时
			err := ctx.Err()
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				_, _ = p.markStatus(t.ID, task.ExecStatusDeadlineExceeded)
			case errors.Is(err, context.Canceled):
				_, _ = p.markStatus(t.ID, task.ExecStatusCancelled)
			}
			end = true
		case res, ok := <-ch:
			if !ok {
				end = true
				break
			}
			switch res.Status {
			case executor.StatusSuccess:
				_, _ = p.markStatus(t.ID, task.ExecStatusSuccess)
				end = true
			case executor.StatusFailed:
				_, _ = p.markStatus(t.ID, task.ExecStatusFailed)
				end = true
			default:
				err := p.executionDAO.UpdateProgress(ctx, eid, uint8(res.Progress))
				if err != nil {
					p.logger.Error("更新任务进度失败",
						slog.Int64("execution_id", eid), slog.Any("progress", res.Progress))
				}
			}
		}
	}
}

func (p *PreemptScheduler) refreshTask(ctx context.Context, ticker *time.Ticker, id int64) error {
	for {
		select {
		case <-ticker.C:
			ctx2, cancel := context.WithTimeout(context.Background(), time.Second*3)
			err := p.dao.UpdateUtime(ctx2, id)
			cancel()
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *PreemptScheduler) releaseTask(t task.Task) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := p.dao.Release(ctx, t)
	if err != nil {
		p.logger.Error("释放任务失败",
			slog.Int64("TaskID", t.ID),
			slog.Any("error", err))
	}
}

func (p *PreemptScheduler) setNextTime(t task.Task) {
	next, err := t.NextTime()
	if err != nil {
		p.logger.Error("计算任务下一次执行时间失败",
			slog.Int64("TaskID", t.ID),
			slog.Any("error", err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	if next.IsZero() {
		err := p.dao.Stop(ctx, t.ID)
		if err != nil {
			p.logger.Error("停止任务调度失败",
				slog.Int64("TaskID", t.ID),
				slog.Any("error", err))
		}
	}
	err = p.dao.UpdateNextTime(ctx, t.ID, next)
	if err != nil {
		p.logger.Error("更新下一次执行时间出错",
			slog.Int64("TaskID", t.ID),
			slog.Any("error", err))
	}
}

func (p *PreemptScheduler) markStatus(tid int64, status task.ExecStatus) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	eid, err := p.executionDAO.InsertExecStatus(ctx, tid, status)
	if err != nil {
		p.logger.Error("记录任务执行失败",
			slog.Int64("task_id", tid),
			slog.Any("error", err))
	}
	if status == task.ExecStatusSuccess {
		err = p.executionDAO.UpdateProgress(ctx, eid, uint8(100))
		if err != nil {
			p.logger.Error("更新任务执行进度为100失败",
				slog.Int64("task_id", tid),
				slog.Int64("execution_id", eid),
				slog.Any("error", err))
		}
	}
	return eid, err
}
