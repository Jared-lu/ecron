package preempter

import (
	"context"
	"errors"
	"github.com/ecodeclub/ecron/internal/storage"
	daomocks "github.com/ecodeclub/ecron/internal/storage/mocks"
	"github.com/ecodeclub/ecron/internal/task"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestDBPreempter_Preempt(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) storage.TaskDAO
		wantTask task.Task
		wantErr  error
	}{
		{
			name: "success",
			mock: func(ctrl *gomock.Controller) storage.TaskDAO {
				dao := daomocks.NewMockTaskDAO(ctrl)
				dao.EXPECT().Preempt(gomock.Any()).Return(task.Task{
					ID: 1,
				}, nil)
				return dao
			},
			wantTask: task.Task{
				ID: 1,
			},
			wantErr: nil,
		},
		{
			name: "failed",
			mock: func(ctrl *gomock.Controller) storage.TaskDAO {
				dao := daomocks.NewMockTaskDAO(ctrl)
				dao.EXPECT().Preempt(gomock.Any()).Return(task.Task{}, errors.New("mock error"))
				return dao
			},
			wantErr: errors.New("mock error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			p := NewDBPreempter(tc.mock(ctrl), time.Second, logger)
			res, cancelFn, err := p.Preempt(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantTask, res)
			assert.NotNil(t, cancelFn)
		})
	}
}
