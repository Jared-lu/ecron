package generic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Caller struct {
	cc      *grpc.ClientConn
	client  *grpcreflect.Client
	file    *desc.FileDescriptor
	service *desc.ServiceDescriptor
	resp    []byte
	err     error
}

func NewCaller(cc *grpc.ClientConn, cli *grpcreflect.Client, file *desc.FileDescriptor, service *desc.ServiceDescriptor) *Caller {
	return &Caller{
		cc:      cc,
		client:  cli,
		file:    file,
		service: service,
	}
}

func (c *Caller) Invoke(ctx context.Context, method string, args interface{}, opts ...grpc.CallOption) *Caller {
	md := c.service.FindMethodByName(method)
	if md == nil {
		// 没有这个方法
		c.err = errors.New("method not found: " + method)
		return c
	}
	invoker, err := c.CreateMethodInvoker(method)
	if err != nil {
		c.err = err
		return c
	}
	reqMessage := invoker.MsgFactory.NewMessage(invoker.Md.GetInputType())
	jsonBytes, err := json.Marshal(args)
	if err != nil {
		c.err = err
		return c
	}
	if err = reqMessage.(*dynamic.Message).UnmarshalJSON(jsonBytes); err != nil {
		c.err = fmt.Errorf("unmarshal req bytes error: %s", err.Error())
		return c
	}
	resp, err := invoker.invoke(ctx, reqMessage, opts...)
	if err != nil {
		c.err = err
		return c
	}
	data, _ := resp.MarshalJSON()
	c.resp = data
	return c
}

func (c *Caller) Result(res any) *Caller {
	er := json.Unmarshal(c.resp, res)
	if er != nil {
		c.err = fmt.Errorf("unmarshal result error: %s", er.Error())
		return c
	}
	return c
}
func (c *Caller) Error() error {
	return c.err
}

func (c *Caller) CreateMethodInvoker(method string) (*Invoker, error) {

	mi := new(Invoker)
	mi.Md = c.service.FindMethodByName(method)
	if mi.Md == nil {
		return nil, fmt.Errorf("method %s not found", method)
	}

	var ext dynamic.ExtensionRegistry
	fmt.Println("print: ", mi.Md.GetInputType())
	if err := c.fetchAllExtensions(&ext, mi.Md.GetInputType()); err != nil {
		return nil, err
	}
	// 同理，这是找出响应message，返回什么给我
	if err := c.fetchAllExtensions(&ext, mi.Md.GetOutputType()); err != nil {
		return nil, err
	}
	mi.MsgFactory = dynamic.NewMessageFactoryWithExtensionRegistry(&ext)

	// stub 客户端stub，和服务器通信用的
	mi.Stub = grpcdynamic.NewStubWithMessageFactory(c.cc, mi.MsgFactory)
	return mi, nil
}

func (c *Caller) fetchAllExtensions(ext *dynamic.ExtensionRegistry, md *desc.MessageDescriptor) (err error) {
	// message name,如 HelloReq
	msgTypeName := md.GetFullyQualifiedName()
	fmt.Println("msgTypeName: ", msgTypeName)
	// 判断message有没有字段
	if len(md.GetExtensionRanges()) > 0 {
		fds, err := c.AllExtensionsForType(msgTypeName)
		if err != nil {
			return fmt.Errorf("failed to query for extensions of type %s: %v", msgTypeName, err)
		}
		// fd => field descriptor
		for _, fd := range fds {
			// 记录这个字段
			if err := ext.AddExtension(fd); err != nil {
				return fmt.Errorf("could not register extension %s of type %s: %v", fd.GetFullyQualifiedName(), msgTypeName, err)
			}
		}
	}
	// recursively fetch extensions for the types of any message fields
	for _, fd := range md.GetFields() {
		if fd.GetMessageType() != nil {
			err := c.fetchAllExtensions(ext, fd.GetMessageType())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Caller) AllExtensionsForType(typeName string) ([]*desc.FieldDescriptor, error) {
	var exts []*desc.FieldDescriptor
	nums, err := c.client.AllExtensionNumbersForType(typeName)
	if err != nil {
		return nil, reflectionSupport(err)
	}

	for _, fieldNum := range nums {
		ext, err := c.client.ResolveExtension(typeName, fieldNum)
		if err != nil {
			return nil, reflectionSupport(err)
		}
		exts = append(exts, ext)
	}

	return exts, nil
}

// Invoker invoke method (Md) with server with Stub
type Invoker struct {
	Md         *desc.MethodDescriptor
	MsgFactory *dynamic.MessageFactory
	Stub       grpcdynamic.Stub
}

func (i *Invoker) invoke(ctx context.Context, request proto.Message, opts ...grpc.CallOption) (resp *dynamic.Message, err error) {
	res, err := i.Stub.InvokeRpc(ctx, i.Md, request, opts...)
	if err != nil {
		return
	}
	if res == nil {
		return
	}
	// 返回响应
	if r, ok := res.(*dynamic.Message); ok {
		resp = r
		return
	}
	return
}

func reflectionSupport(err error) error {
	if err == nil {
		return nil
	}

	if stat, ok := status.FromError(err); ok && stat.Code() == codes.Unimplemented {
		return errors.New("不支持反射")
	}

	return err
}
