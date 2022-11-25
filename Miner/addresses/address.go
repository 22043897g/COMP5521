package addresses

import (
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
)

var rdb *redis.Client

func initRedis() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // 指定
		Password: "",
		DB:       0, // redis一共16个库，指定其中一个库即可
	})
	_, err = rdb.Ping().Result()
	return
}

func RedisInit() {
	err := initRedis()
	if err != nil {
		fmt.Printf("connect redis failed! err : %v\n", err)
		return
	}
	//fmt.Println("redis连接成功！")
}

func SaveNewAddress(address string) {
	RedisInit()
	//因为地址数组下标从0开始，所以总数正好是存储新地址的索引值
	res, _ := rdb.Get("AddressAmount").Result()
	key := "A" + res
	println(key)
	rdb.Set(key, address, 0)
	//地址总数+1后存回
	num, _ := strconv.ParseInt(res, 10, 64)
	newnum := num + 1
	value := strconv.FormatInt(newnum, 10)
	rdb.Set("AddressAmount", value, 0)
	println("New address:", address, " saved.")
}

func ReadAllAddress() []string {
	RedisInit()
	addrs := []string{}
	res, _ := rdb.Get("AddressAmount").Result()
	println(res)
	num, _ := strconv.ParseInt(res, 10, 64)
	for i := 0; i < int(num); i++ {
		key := "A" + strconv.FormatInt(int64(i), 10)
		addr, _ := rdb.Get(key).Result()
		addrs = append(addrs, addr)
	}
	return addrs
}

// 查询是否记录了某地址
func CheckAddress(address string) bool {
	RedisInit()
	res, _ := rdb.Get("AddressAmount").Result()
	num, _ := strconv.ParseInt(res, 10, 64)
	for i := 0; i < int(num); i++ {
		key := "A" + strconv.FormatInt(int64(i), 10)
		addr, _ := rdb.Get(key).Result()
		if addr == address {
			return true
		}
	}
	return false
}

// PortInit 清空除了miner以外所有端口地址
func PortInit() {
	RedisInit()
	rdb.Set("AddressAmount", 1, 0)
	rdb.Set("A0", "0.0.0.0:8001", 0)
	rdb.Del("A1")
	rdb.Del("A2")
	rdb.Del("A3")
}
