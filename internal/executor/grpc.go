package executor

import (
	"context"
	"encoding/json"
	"github.com/ecodeclub/ecron/internal/errs"
	"github.com/ecodeclub/ecron/internal/task"
	"github.com/ecodeclub/ecron/pkg/generic"
	"google.golang.org/grpc"
	"log/slog"
	"time"
)

type GrpcExecutor struct {
	logger      *slog.Logger
	callerCache map[string]*generic.Caller
	opts        []grpc.DialOption
}

func NewGrpcExecutor() Executor {
	return &GrpcExecutor{}
}

func (g *GrpcExecutor) Name() string {
	return "GRPC"
}

func (g *GrpcExecutor) Run(ctx context.Context, t task.Task, eid int64) (task.ExecStatus, error) {
	cfg, err := g.parseCfg(t.Cfg)
	if err != nil {
		g.logger.Error("任务配置信息错误",
			slog.Int64("task_id", t.ID), slog.String("configuration", t.Cfg))
		return task.ExecStatusFailed, errs.ErrInCorrectConfig
	}

	var caller *generic.Caller
	if c, ok := g.callerCache[cfg.ServiceName]; !ok || c == nil {
		conn, err := grpc.NewClient(cfg.Target, g.opts...)
		if err != nil {
			panic(err)
		}
		client := generic.NewGeneticClient(conn)
		caller, err = client.FindService(ctx, cfg.ServiceName)
		if err != nil {
			// 不支持泛化调用
			return task.ExecStatusFailed, err
		}
	}
	caller = g.callerCache[cfg.ServiceName]

	req := request{
		Eid:    eid,
		Action: actionExecute,
	}
	var result Result
	err = caller.Invoke(ctx, cfg.Method, &req).Result(&result).Error()
	if err != nil {
		return task.ExecStatusFailed, err
	}

	switch result.Status {
	case StatusSuccess:
		return task.ExecStatusSuccess, nil
	case StatusRunning:
		return task.ExecStatusRunning, nil
	default:
		return task.ExecStatusFailed, nil
	}
}

func (g *GrpcExecutor) Explore(ctx context.Context, eid int64, t task.Task) <-chan Result {
	resultChan := make(chan Result, 1)
	go g.explore(ctx, resultChan, t, eid)
	return resultChan
}

func (g *GrpcExecutor) explore(ctx context.Context, ch chan Result, t task.Task, eid int64) {

}

func (g *GrpcExecutor) TaskTimeout(t task.Task) time.Duration {
	result, err := g.parseCfg(t.Cfg)
	if err != nil || result.TaskTimeout < 0 {
		return time.Minute
	}
	return result.TaskTimeout
}

func (g *GrpcExecutor) parseCfg(cfg string) (GrpcCfg, error) {
	var result GrpcCfg
	err := json.Unmarshal([]byte(cfg), &result)
	return result, err
}

type GrpcCfg struct {
	Target      string `json:"target"`
	ServiceName string `json:"service_name"`
	Method      string `json:"method"`
	// 预计任务执行时长
	TaskTimeout time.Duration `json:"taskTimeout"`
	// 任务探查间隔
	ExploreInterval time.Duration `json:"exploreInterval"`
}

type request struct {
	Eid    int64  `json:"eid"`
	Action action `json:"action"`
}

type action int

const (
	actionUnknown action = iota
	actionExecute
	actionExplore
	actionStop
)
