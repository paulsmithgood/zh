package rpc

import (
	"context"
	"log"
	"network/protoc/gen"
	"testing"
	"time"
)

type UserService struct {
	// 用反射来赋值
	// 类型是函数的字段，它不是方法（它不是定义在 UserService 上的方法）
	// 本质上是一个字段
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)

	GetByIdProto func(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error)
}

func (u UserService) Name() string {
	return "user-service"
}

type GetByIdReq struct {
	Id int
}

type GetByIdResp struct {
	Msg string
}

type UserServiceServer struct {
	Err error
	Msg string
}

func (u *UserServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	log.Println(req)
	return &GetByIdResp{
		Msg: u.Msg,
	}, u.Err
}

func (u *UserServiceServer) GetByIdProto(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	log.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: u.Msg,
		},
	}, u.Err
}

func (u *UserServiceServer) Name() string {
	return "user-service"
}

type UserServiceServerTIMEOUT struct {
	t     *testing.T
	sleep time.Duration
	Err   error
	Msg   string
}

func (u *UserServiceServerTIMEOUT) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	if _, ok := ctx.Deadline(); !ok {
		//panic("要设置超时时间")
		//fmt.Println("没有设置超时")
		u.t.Fatal("要设置超时时间000")
		//fmt.Println("要设置超时时间000")
	}
	time.Sleep(u.sleep)
	return &GetByIdResp{
		Msg: u.Msg,
	}, u.Err
}
func (u *UserServiceServerTIMEOUT) Name() string {
	return "user-service"
}
