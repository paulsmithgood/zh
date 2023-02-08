package cache

//func TestNewClient(t *testing.T) {
//	rdb := redis.NewClient(&redis.Options{
//		Addr: "localhost:50055",
//	})
//	_ = NewClient(rdb)
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
//	res, err := rdb.Set(ctx, "key", "value1", time.Minute*5).Result()
//	fmt.Println(res, err)
//	cancel()
//}
//
//func TestClient_Lock(t *testing.T) {
//	rdb := redis.NewClient(&redis.Options{
//		Addr: "localhost:50055",
//	})
//	c := NewClient(rdb)
//	//c.Set()
//
//	//time.Sleep(time.Second * 10)
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
//	fix := &FixedIntervalRetryStrategy{Interval: time.Second * 2, MaxCnt: 5}
//	l, err := c.Lock(ctx, "key", time.Second*30, time.Second*2, fix)
//	fmt.Println(l, err)
//	cancel()
//}
