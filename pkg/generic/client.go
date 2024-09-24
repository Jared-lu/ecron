package generic

import (
	"context"
	"errors"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type GeneticClient struct {
	cc grpc.ClientConnInterface
}

func NewGeneticClient(cc grpc.ClientConnInterface) *GeneticClient {
	return &GeneticClient{
		cc: cc,
	}
}

func (g GeneticClient) FindService(ctx context.Context, symbol string) (*Caller, error) {
	cli := grpcreflect.NewClientV1Alpha(ctx,
		grpc_reflection_v1alpha.NewServerReflectionClient(g.cc))
	// 向grpc server寻找对应服务的文件描述符，使用给定的完全限定名
	file, err := cli.FileContainingSymbol(symbol)
	if err != nil {
		// 不支持反射服务，也就是对方没有注册反射服务
		return nil, err
	}
	// 找出指定symbol的描述符
	d := file.FindSymbol(symbol)
	if d == nil {
		// 根本就没有这个服务
		return nil, errors.New("symbol not found")
	}

	sd, ok := d.(*desc.ServiceDescriptor)
	if !ok {
		// 不支持泛化调用
		return nil, errors.New("service not found")
	}
	return NewCaller(g.cc.(*grpc.ClientConn), cli, file, sd), nil
}
