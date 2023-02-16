package rpc

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"network/protoc/gen"
	"network/rpc/compression/gzip"
	"network/rpc/serialize/proto"
	"testing"
	"time"
)

func TestInitServiceProto(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	server.RegisterSerializer(&proto.Serializer{})
	server.RegisterCompression(&gzip.Compression{})
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	usClient := &UserService{}
	client, err := NewClient(":8081", ClientWithSerializer(&proto.Serializer{}), ClientWithCompression(&gzip.Compression{}))
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct {
		name string
		mock func()

		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello, world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("mock error")
			},
			wantResp: &GetByIdResp{},
			wantErr:  errors.New("mock error"),
		},

		{
			name: "both",
			mock: func() {
				service.Msg = "hello, world"
				service.Err = errors.New("mock error")
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
			wantErr: errors.New("mock error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, er := usClient.GetByIdProto(context.Background(), &gen.GetByIdReq{Id: 123})

			//异步
			//var respAsync *gen.GetByIdResp
			//var wg sync.WaitGroup
			//wg.Add(1)
			//go func() {
			//	respAsync, err = usClient.GetByIdProto(context.Background(), &gen.GetByIdReq{Id: 123})
			//	wg.Done()
			//}()
			////比如在这里做了很多事情
			//
			//wg.Wait()

			//回调
			//go func() {
			//	respAsync, err := usClient.GetByIdProto(context.Background(), &gen.GetByIdReq{Id: 123})
			//	//随便怎么处理
			//	//respAsync.User.
			//}()

			//虚假单项调用  ：他其实已经返回了响应，只是我们吧响应给忽略掉了
			//go func() {
			//	_, _ = usClient.GetByIdProto(context.Background(), &gen.GetByIdReq{Id: 123})
			//}()
			assert.Equal(t, tc.wantErr, er)
			if resp != nil && resp.User != nil {
				assert.Equal(t, tc.wantResp.Msg, resp.User.Name)
			}
		})
	}
}

func TestInitClientProxy(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	usClient := &UserService{}
	client, err := NewClient(":8081")
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct {
		name string
		mock func()

		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello, world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("mock error")
			},
			wantResp: &GetByIdResp{},
			wantErr:  errors.New("mock error"),
		},

		{
			name: "both",
			mock: func() {
				service.Msg = "hello, world"
				service.Err = errors.New("mock error")
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
			wantErr: errors.New("mock error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			//resp, _ := usClient.GetById(CtxWithOneway(context.Background()), &GetByIdReq{})
			//这样的话就不能用resp 了，因为他没有任何返回

			resp, er := usClient.GetById(context.Background(), &GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestOneway(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	usClient := &UserService{}
	client, err := NewClient(":8081")
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct {
		name string
		mock func()

		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "oneway",
			mock: func() {
				service.Err = errors.New("MOCK ERROR")
				service.Msg = "hello, world"
			},
			wantResp: &GetByIdResp{},
			wantErr:  errors.New("这是一个oneway调用..."),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			//resp, _ := usClient.GetById(CtxWithOneway(context.Background()), &GetByIdReq{})
			//这样的话就不能用resp 了，因为他没有任何返回
			ctx := CtxWithOneway(context.Background())
			resp, er := usClient.GetById(ctx, &GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
			//time.Sleep(time.Second*2)
			//assert.Equal(t, "hello,world",service.Msg)
		})
	}
}

func TestTimeout(t *testing.T) {
	server := NewServer()
	service := &UserServiceServerTIMEOUT{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	usClient := &UserService{}
	client, err := NewClient(":8081")
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct {
		name string
		mock func() context.Context

		wantErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "timeout",
			mock: func() context.Context {
				service.t = t
				service.Err = errors.New("MOCK ERROR")
				service.Msg = "hello, world"
				service.sleep = time.Second * 2
				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				return ctx
			},
			wantResp: &GetByIdResp{},
			wantErr:  context.DeadlineExceeded,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//tc.mock()
			//resp, _ := usClient.GetById(CtxWithOneway(context.Background()), &GetByIdReq{})
			//这样的话就不能用resp 了，因为他没有任何返回
			//ctx := CtxWithOneway(context.Background())
			resp, er := usClient.GetById(tc.mock(), &GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
			//time.Sleep(time.Second*2)
			//assert.Equal(t, "hello,world",service.Msg)
		})
	}
}
