package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ecodeclub/ecron/internal/errs"
	"github.com/ecodeclub/ecron/internal/task"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type HttpExecutor struct {
	logger *slog.Logger
	client *http.Client
	// 任务探查最大失败次数
	maxFailCount int
}

func NewHttpExecutor(logger *slog.Logger) *HttpExecutor {
	return &HttpExecutor{
		client: &http.Client{
			// http调用超时配置
			Timeout: time.Second * 5,
		},
		logger:       logger,
		maxFailCount: 5,
	}
}

func (h *HttpExecutor) Name() string {
	return "HTTP"
}

func (h *HttpExecutor) Run(ctx context.Context, t task.Task, eid int64) (task.ExecStatus, error) {
	cfg, err := h.parseCfg(t.Cfg)
	if err != nil {
		h.logger.Error("任务配置信息错误",
			slog.Int64("ID", t.ID), slog.String("Cfg", t.Cfg))
		return task.ExecStatusFailed, errs.ErrInCorrectConfig
	}

	request, err := http.NewRequest(cfg.Method, cfg.Url, bytes.NewBuffer([]byte(cfg.Body)))
	if err != nil {
		return task.ExecStatusFailed, err
	}
	if cfg.Header == nil {
		request.Header = make(http.Header)
	} else {
		request.Header = cfg.Header
	}
	request.Header.Set("execution_id", fmt.Sprintf("%v", eid))
	resp, err := h.client.Do(request)
	ok := os.IsTimeout(err)
	if ok {
		return task.ExecStatusDeadlineExceeded, errs.ErrRequestFailed
	}
	if err != nil {
		h.logger.Error("发起任务执行请求失败",
			slog.Int64("ID", t.ID), slog.Any("error", err))
		return task.ExecStatusFailed, errs.ErrRequestFailed
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return task.ExecStatusFailed, nil
	}

	var result Result
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return task.ExecStatusFailed, err
	}
	switch result.Status {
	case StatusSuccess:
		return task.ExecStatusSuccess, nil
	case StatusFailed:
		return task.ExecStatusFailed, nil
	case StatusRunning:
		return task.ExecStatusRunning, nil
	default:
		return task.ExecStatusFailed, nil
	}
}

func (h *HttpExecutor) Explore(ctx context.Context, eid int64, t task.Task) <-chan Result {
	resultChan := make(chan Result, 1)
	go h.explore(ctx, resultChan, t, eid)
	return resultChan
}

func (h *HttpExecutor) explore(ctx context.Context, ch chan Result, t task.Task, eid int64) {
	defer close(ch)

	cfg, _ := h.parseCfg(t.Cfg)
	failResult := Result{
		Eid:    eid,
		Status: StatusFailed,
	}

	failCount := 0
	ticker := time.NewTicker(cfg.ExploreInterval)

	for {
		select {
		case <-ctx.Done():
			// 通知业务方取消任务执行
			h.cancelExec(cfg, eid)
			return
		case <-ticker.C:
			request, err := http.NewRequest(cfg.Method, cfg.Url, bytes.NewBuffer([]byte(cfg.Body)))
			if err != nil {
				if failCount >= h.maxFailCount {
					ch <- failResult
					return
				}
				failCount++
			}
			if cfg.Header == nil {
				request.Header = make(http.Header)
			} else {
				request.Header = cfg.Header
			}
			request.Header.Set("execution_id", fmt.Sprintf("%v", eid))
			resp, err := h.client.Do(request)
			if err != nil || resp.StatusCode != http.StatusOK {
				if failCount >= h.maxFailCount {
					ch <- failResult
					return
				}
				failCount++
			}
			_ = resp.Body.Close()
			var result Result
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				if failCount >= h.maxFailCount {
					ch <- failResult
					return
				}
				failCount++
			}
			ch <- result
			if result.Status != StatusRunning {
				return
			}
		default:
		}
	}
}

func (h *HttpExecutor) TaskTimeout(t task.Task) time.Duration {
	result, err := h.parseCfg(t.Cfg)
	if err != nil || result.TaskTimeout < 0 {
		return time.Minute
	}
	return result.TaskTimeout
}

func (h *HttpExecutor) parseCfg(cfg string) (HttpCfg, error) {
	var result HttpCfg
	err := json.Unmarshal([]byte(cfg), &result)
	return result, err
}

func (h *HttpExecutor) parseResult(body []byte) (Result, error) {
	var result Result
	err := json.Unmarshal(body, &result)
	return result, err
}

func (h *HttpExecutor) cancelExec(cfg HttpCfg, eid int64) {
	request, err := http.NewRequest(cfg.Method, cfg.Url, bytes.NewBuffer([]byte(cfg.Body)))
	if err != nil {
		h.logger.Error("通知业务方停止执行任务失败", slog.Int64("execution_id", eid))
		return
	}
	if cfg.Header == nil {
		request.Header = make(http.Header)
	} else {
		request.Header = cfg.Header
	}
	request.Header.Set("execution_id", fmt.Sprintf("%v", eid))
	// 当业务方收到带有 Header cancel=true 的请求时，要停止执行任务
	request.Header.Set("cancel", fmt.Sprintf("%v", true))
	resp, err := h.client.Do(request)
	if err != nil || resp.StatusCode != http.StatusOK {
		h.logger.Error("通知业务方停止执行任务失败", slog.Int64("execution_id", eid))
	}
}

type HttpCfg struct {
	Method string      `json:"method"`
	Url    string      `json:"url"`
	Header http.Header `json:"header"`
	Body   string      `json:"body"`
	// 预计任务执行时长
	TaskTimeout time.Duration `json:"taskTimeout"`
	// 任务探查间隔
	ExploreInterval time.Duration `json:"exploreInterval"`
}
