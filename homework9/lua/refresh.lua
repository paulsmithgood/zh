--判断value 是不是我的 然后增加要刷新的时间

if redis.call('get',KEY[1])==ARGV[1] then
    --说明是我的value
    return redis.call('EXPIRE',KEY[1],ARGV[2])
else
    --不是我的
    return 0
end

--127.0.0.1:6379> expire key 30
--(integer) 1