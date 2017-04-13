# MASS

[![Build Status](https://travis-ci.org/qianlnk/mass.svg?branch=master)](https://travis-ci.org/qianlnk/mass)

用用户提供的`processing method`生产ID，生产ID时确保在本次生产的ID完成前同一组参数不再生产其它不一致的ID（解决并发、重试、刷机、多机部署等问题），另外查库插库需要用户在`processing method`中自己实现，以确保所生产的ID参数在数据库中不重复。

## The main technical stack

* redis实现分布式锁
* redis订阅发布
* 生产频率限制

## API

`processing method`

```golang
type ProcessingMethod func(materials ...interface{}) interface{}
```

## How to use limit

```golang
func TestLimiter(t *testing.T) {
    l := NewLimiter(3)

    for {
        l.Limit()
        fmt.Println("hello qianlnk")
    }
}
```

## How to use lock

```golang
rc := pool.Get()
ok, err := Lock(rc, "testkey", "testsecret", 20)
if ok {
    //do sth with lock
}else {
    //get lock failed
}

Unlock(rc, "testkey", "testsecret")
```

## Usage

set redis

```golang
func StartFactory(redisHost string, redisDB int, redisMaxIdle int, redisMaxActive int){}
```

```golang
mass.StartFactory("127.0.0.1:6379", 2, 100, 1000)
```

create an new process, params is product name, method, timeout, args...

```golang
func NewProduct(name string, method ProcessingMethod, timeout int, materials ...interface{}) Forklift {}
```

```golang
p := mass.NewProduct(strconv.Itoa(i), howToProcessing, 1, i)
```

and, get result

```golang
func (f Forklift) Get(v interface{}) {}
```

```golang
var test testRes
p.Get(&test)
```

see demo for more info.