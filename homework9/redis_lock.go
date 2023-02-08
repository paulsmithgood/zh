package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

//docker exec -it myredis redis-cli
//redis lock 分布式锁

var (
	Lock_Failtopreempt = errors.New("锁已经被其他人抢占")
	Lock_NotHoldLock   = errors.New("锁不是我持有的")
	//go:embed lua/unlock.lua
	unlocklua string
	//go:embed lua/refresh.lua
	Refreshlua string
	//go:embed lua/lock.lua
	locklua string
)

type client struct {
	c redis.Cmdable
}

func NewClient(cmdable redis.Cmdable) *client {
	return &client{c: cmdable}
}

// 如果存在就不动
func (c *client) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	//ctx 用来控制调用超时
	value := uuid.New().String()
	ok, err := c.c.SetNX(ctx, key, value, expiration).Result()
	//SET if Not eXists
	//set
	if err != nil {
		//	说明在setnx 里面 可能出现了error?
		return nil, err
	}
	if !ok {
		//说明没有被替换成功 可能锁已经被人抢了
		return nil, Lock_Failtopreempt
	}
	//到这里说明 锁已经被人抢成功了
	return &Lock{
		client:     c.c,
		key:        key,
		value:      value,
		expiration: expiration,
		unlockchan: make(chan struct{}, 1),
	}, nil
}

func (c *client) Lock(ctx context.Context, key string, expiration time.Duration, timeout time.Duration, retry RetryStrategy) (*Lock, error) {
	//设置一个定时器
	var timer *time.Timer
	val := uuid.New().String()
	for {
		tctx, cancel := context.WithTimeout(ctx, timeout) //一个有超时的
		res, err := c.c.Eval(tctx, locklua, []string{key}, val, expiration).Result()
		//fmt.Println(res, err)
		cancel()
		if err != nil && nil != context.DeadlineExceeded {
			//redis 是出现的其他问题返回 直接return
			return nil, err
		}
		if res == "OK" {
			//成功了
			return &Lock{
				client:     c.c,
				key:        key,
				value:      val,
				expiration: expiration,
				unlockchan: make(chan struct{}, 1),
			}, err
		}

		interval, ok := retry.Next()
		fmt.Println("开始重试")
		if !ok {
			return nil, errors.New("超过了重试次数了--")
		}
		if timer == nil {
			timer = time.NewTimer(interval)
		} else {
			timer.Reset(interval)
		}

		select {
		//等待定时器的信号
		case <-timer.C:
		//等待超时
		case <-ctx.Done():
			return nil, ctx.Err()
		}

	}

}

func (c *client) Set() {
	//test
	c.c.Set(context.Background(), "key", "xxx", time.Second*60)
}

type Lock struct {
	client     redis.Cmdable
	key        string
	value      string
	expiration time.Duration
	unlockchan chan struct{}
}

func (l *Lock) Unlock(ctx context.Context) error {
	//解锁 需要去验证 在redis 中通过key 拿到的 value 是不是 我们的value
	res, err := l.client.Eval(ctx, unlocklua, []string{l.key}, l.value).Int64()
	defer func() {
		select {
		case l.unlockchan <- struct{}{}:
		default:
			//说明没有人调用
		}
	}()
	if err == redis.Nil {
		//说明 我们 的key 没有被持有 在get 下就没有拿到结果
		return Lock_NotHoldLock
	}
	if err != nil {
		//说明这个是redis 里面的错误
		return err
	}
	if res != 1 {
		//说明 del 返回的结果不是1 说明锁已经被别人拿走了
		return Lock_NotHoldLock
	}
	return nil //说明解锁成功了
}

func (l *Lock) Refresh(ctx context.Context) error {
	//刷新前也要判断一下这个value 是不是我的
	res, err := l.client.Eval(ctx, Refreshlua, []string{l.key}, l.value, l.expiration.Seconds()).Int64()
	if err == redis.Nil {
		//说明都没有get到key
		return Lock_NotHoldLock
	}
	if err != nil {
		//这里的肯定是redis 里面出了问题 context.DeadlineExceeded
		return err
	}
	if res != 1 {
		//说明get 到的key 不是我的
		return Lock_NotHoldLock
	}
	return nil //说明续约已经成功了
}

func (l *Lock) AutoRefresh(interval time.Duration, timeout time.Duration) error {
	//interval 调用 refresh 的间隔时间,
	ticker := time.NewTicker(interval)

	timeoutchan := make(chan struct{}, 1)
	for {
		select {
		case <-ticker.C:
			//收到了间隔调用的信号
			ctx_, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx_)
			cancel()

			if err == context.DeadlineExceeded {
				timeoutchan <- struct{}{}
				continue
			}
		case <-timeoutchan:
			//收到了 timeout 发来的信号
			ctx_, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx_)
			cancel()

			if err == context.DeadlineExceeded {
				timeoutchan <- struct{}{}
				continue
			}
		case <-l.unlockchan:
			return nil
		}
	}
}