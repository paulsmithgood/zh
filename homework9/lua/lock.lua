
local val=redis.call('get',KEYS[1])
if val==false then
    --没有拿到key
    return redis.call('set',KEYS[1],ARGV[1],'EX',ARGV[2])

    --redis> SET anotherkey "will expire in a minute" EX 60
    --"OK"
elseif val==ARGV[1] then
    --说明 有这个key 并且是我自己的key
    redis.call('expire',KEYS[1],ARGV[2])
    return 'OK'
else
    --有key 但是不是我自己的
    return ''
end
