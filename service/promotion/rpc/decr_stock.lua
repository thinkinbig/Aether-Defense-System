-- KEYS[1]: Inventory Key
-- ARGV[1]: Deduction Quantity

local stock = redis.call('GET', KEYS[1])

if (stock == false) then
    return {err = "Inventory Key does not exist"}
end

if (tonumber(stock) < tonumber(ARGV[1])) then
    return {err = "Insufficient inventory"}
end

redis.call('DECRBY', KEYS[1], ARGV[1])
return {result = "Inventory deduction successful"}
