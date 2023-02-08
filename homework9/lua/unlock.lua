if redis.call('get',KEYS[1])==ARGV[1] then
    return redis.call('del',KEYS[1])
else
    return 0
end


--redis> SET key1 "Hello"
--"OK"
--redis> SET key2 "World"
--"OK"
--redis> DEL key1 key2 key3
--(integer) 2
--redis>
