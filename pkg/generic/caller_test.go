package generic

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"
)

func TestCaller_Invoke(t *testing.T) {
	conn, err := grpc.NewClient("127.0.0.1:8888", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := NewGeneticClient(conn)
	// 生成一client，这个客户端要初始化，
	caller, err := client.FindService(ctx, "Hello")
	require.NoError(t, err)
	req := map[string]interface{}{"uid": "1001", "message": "json"}
	var result HelloResp
	err = caller.Invoke(ctx, "SayHello", &req).Result(&result).Error()
	require.NoError(t, err)
	t.Log(result)
}

type HelloResp struct {
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}
