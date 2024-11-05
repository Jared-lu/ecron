package mysql

import (
	"github.com/ecodeclub/ecron/internal/task"
	"time"
)

type TaskInfo struct {
	ID   int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name string `gorm:"column:name"`
	// 任务类型
	Type         string `gorm:"column:type"`
	Cron         string `gorm:"column:cron"`
	Executor     string `gorm:"column:executor"`
	Owner        string `gorm:"column:owner"`
	Status       int8   `gorm:"column:status;index:idx_status_utime;index:idx_status_next_exec_time"`
	Cfg          string `gorm:"column:cfg"`
	NextExecTime int64  `gorm:"column:next_exec_time;index:idx_status_next_exec_time"`
	Ctime        int64  `gorm:"column:ctime"`
	Utime        int64  `gorm:"column:utime;index:idx_status_utime;"`
}

func (TaskInfo) TableName() string {
	return "task_info"
}

func toEntity(t task.Task) TaskInfo {
	return TaskInfo{
		ID:       t.ID,
		Name:     t.Name,
		Type:     t.Type.String(),
		Cron:     t.CronExp,
		Executor: t.Executor,
		Cfg:      t.Cfg,
		Ctime:    t.Ctime.UnixMilli(),
		Utime:    t.Utime.UnixMilli(),
		Owner:    t.Owner,
	}
}

func toTask(t TaskInfo) task.Task {
	return task.Task{
		ID:         t.ID,
		Name:       t.Name,
		Type:       task.Type(t.Type),
		Executor:   t.Executor,
		Cfg:        t.Cfg,
		CronExp:    t.Cron,
		Ctime:      time.UnixMilli(t.Ctime),
		Utime:      time.UnixMilli(t.Utime),
		LastStatus: t.Status,
		Owner:      t.Owner,
	}
}

// Execution 任务执行记录
type Execution struct {
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	// 一个任务至多一个执行记录
	Tid int64 `gorm:"column:tid;uniqueIndex:idx_tid"`
	// 任务执行进度
	Progress uint8 `gorm:"column:progress"`
	// 任务执行状态，0-未知，1-运行中，2-成功，3-失败，4-超时，5-主动取消
	Status uint8 `gorm:"column:status"`
	Ctime  int64 `gorm:"column:ctime"`
	Utime  int64 `gorm:"column:utime"`
}

func (Execution) TableName() string {
	return "execution"
}
