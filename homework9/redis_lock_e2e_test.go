package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClient_e2e_Lock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:50055",
	})
	//lock 会分成几种情况，
	//1.我进去就拿到了锁
	//2.我进来重试了很多次，返回了
	//3.拿到了锁，不过是重试了几次之后
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		key        string
		expiration time.Duration
		timeout    time.Duration
		retry      RetryStrategy

		wantlock  *Lock
		wanterror error
	}{
		{
			name: "locked",
			before: func(t *testing.T) {
				//进去就拿到了锁 就不需要了
			},
			after: func(t *testing.T) {
				//这里用于做 执行完成后的验证
				//查看我们加入的key 判断一下他的剩余时间？
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
				defer cancel()
				lasttime, err := rdb.TTL(ctx, "e2e_locked").Result()
				assert.NoError(t, err)
				//比较lasttime 剩余的时间是不是大于我们预设的

				assert.True(t, lasttime > time.Second*50)

				//比较完成，删除掉
				_, err = rdb.Del(ctx, "e2e_locked").Result()
				assert.NoError(t, err)
				fmt.Println("校验结束-e2e_locked")
			},
			key:        "e2e_locked",
			expiration: time.Minute,
			timeout:    time.Second * 2,
			retry: &FixedIntervalRetryStrategy{
				Interval: time.Second,
				MaxCnt:   10,
			},
			wantlock: &Lock{
				key:        "e2e_locked",
				expiration: time.Minute,
			},
		},
		{
			name: "锁被他人持有",

			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				res, err := rdb.Set(ctx, "e2e_locked2", "xxx", time.Minute*5).Result()
				assert.NoError(t, err)
				assert.Equal(t, res, "OK")
			},
			after: func(t *testing.T) {
				//这里用于做 执行完成后的验证
				//校验一下 key 是否存在并且把他删除
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				res, err := rdb.GetDel(ctx, "e2e_locked2").Result()
				assert.NoError(t, err)
				assert.Equal(t, res, "xxx")
				fmt.Println("校验结束-e2e_locked2")
			},
			key:        "e2e_locked2",
			expiration: time.Minute,
			timeout:    time.Second * 2,
			retry: &FixedIntervalRetryStrategy{
				Interval: time.Second,
				MaxCnt:   10,
			},
			wanterror: errors.New("超过了重试次数了--"),
		},
		{
			name: "拿到了锁，不过是重试了之后",

			before: func(t *testing.T) {
				//在这里要设置一下有比较短的过期时间，可以在过期后可以刚好被重试引用到
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				res, err := rdb.Set(ctx, "e2e_locked3", "xxx", time.Second*10).Result()
				assert.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				//ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				//defer cancel()
				//res, err := rdb.GetDel(ctx, "e2e_lock3", "xxx", time.Second*10).Result()
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				timeout, err := rdb.TTL(ctx, "e2e_locked3").Result()
				assert.NoError(t, err)
				assert.True(t, timeout >= time.Second*50)
				_, err = rdb.Del(ctx, "e2e_locked3").Result()
				assert.NoError(t, err)
			},
			key:        "e2e_locked3",
			expiration: time.Minute,
			timeout:    time.Second,
			retry: &FixedIntervalRetryStrategy{
				Interval: time.Second * 2,
				MaxCnt:   10,
			},
			wantlock: &Lock{
				key:        "e2e_locked3",
				expiration: time.Minute,
			},
		},
	}

	c := NewClient(rdb)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t) //准备数据
			l, err := c.Lock(context.Background(), tc.key, tc.expiration, tc.timeout, tc.retry)
			assert.Equal(t, err, tc.wanterror)
			if err != nil {
				return
			}
			assert.Equal(t, l.key, tc.key)
			assert.Equal(t, l.expiration, tc.expiration)
			assert.NotEmpty(t, l.value)
			assert.NotNil(t, l.client)
			tc.after(t) //校验等
		})
	}
}
