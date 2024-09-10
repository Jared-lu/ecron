package preempter

import (
	"context"
	"errors"
	"github.com/ecodeclub/ecron/internal/errs"
	"github.com/ecodeclub/ecron/internal/storage"
	"github.com/ecodeclub/ecron/internal/task"
	"log/slog"
	"time"
)

type DBPreempter struct {
	dao             storage.TaskDAO
	refreshInterval time.Duration
	logger          *slog.Logger
}

func NewDBPreempter(dao storage.TaskDAO, refreshInterval time.Duration, logger *slog.Logger) *DBPreempter {
	return &DBPreempter{
		dao:             dao,
		refreshInterval: refreshInterval,
		logger:          logger,
	}
}

func (p *DBPreempter) Preempt(ctx context.Context) (task.Task, CancelFn, error) {
	t, err := p.dao.Preempt(ctx)
	if err != nil {
		return task.Task{}, nil, err
	}

	ticker := time.NewTicker(p.refreshInterval)
	go func() {
		for range ticker.C {
			p.logger.Debug("执行任务续约", slog.Int64("task_id", t.ID))
			err := p.refresh(t)
			if err == nil {
				ticker.Reset(p.refreshInterval)
				continue
			} else if errors.Is(err, errs.ErrTaskNotFound) {
				p.logger.Error("任务续约失败，且当前任务已被其它节点抢占", slog.Int64("task_id", t.ID))
				return
			}
			// 调小续约频率，直到成功后恢复正常或者退出
			ticker.Reset(p.refreshInterval / 3)
		}
	}()

	cancelFn := func() {
		p.logger.Info("释放任务", slog.Int64("task_id", t.ID))
		defer ticker.Stop()
		go p.setNextTime(t)
		go p.release(t)
	}
	return t, cancelFn, nil
}

func (p *DBPreempter) refresh(t task.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return p.dao.UpdateUtime(ctx, t)
}

func (p *DBPreempter) release(t task.Task) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := p.dao.Release(ctx, t)
	if err != nil {
		p.logger.Error("释放任务失败",
			slog.Int64("task_id", t.ID), slog.Any("error", err))
	}
	p.logger.Debug("释放任务成功", slog.Int64("task_id", t.ID))
}

func (p *DBPreempter) setNextTime(t task.Task) {
	next, err := t.NextTime()
	if err != nil {
		p.logger.Error("计算任务下一次执行时间失败",
			slog.Int64("task_id", t.ID), slog.String("corn", t.CronExp),
			slog.Any("error", err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	if next.IsZero() {
		err := p.dao.Stop(ctx, t.ID)
		if err != nil {
			p.logger.Error("停止任务调度失败", slog.String("corn", t.CronExp),
				slog.Int64("task_id", t.ID), slog.Any("error", err))
		}
		return
	}
	err = p.dao.UpdateNextTime(ctx, t.ID, next)
	if err != nil {
		p.logger.Error("更新下一次执行时间出错", slog.String("corn", t.CronExp),
			slog.Int64("task_id", t.ID), slog.Any("error", err))
	}
	p.logger.Debug("任务下一次执行时间", slog.Int64("task_id", t.ID), slog.Any("time", next))
}
