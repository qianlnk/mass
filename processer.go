package mass

import (
	"sync"
	"time"

	"fmt"

	"github.com/garyburd/redigo/redis"
)

//ProcessingMethod the processing method
type ProcessingMethod func(materials ...interface{}) interface{}

type Forklift chan interface{}

type ProcessingPool struct {
	productName string
	method      ProcessingMethod
	materials   []interface{}
}

type Product struct {
	name      string
	forklifts []Forklift
}

type Factory struct {
	mu             sync.Mutex
	products       map[string]Product
	processingPool chan ProcessingPool
	maxPool        int
	importPool     *redis.Pool
}

var (
	factory *Factory
	once    sync.Once
)

func StartFactory(redisHost string, redisDB int, redisMaxIdle int, redisMaxActive int) {
	rp := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisHost)
			if err != nil {
				return nil, err
			}

			c.Do("SELECT", redisDB)
			return c, err
		},

		MaxIdle:     redisMaxIdle,
		MaxActive:   redisMaxActive,
		IdleTimeout: time.Second * 1,
	}

	factory = &Factory{
		products:       make(map[string]Product),
		processingPool: make(chan ProcessingPool, 10),
		maxPool:        100,
		importPool:     rp,
	}

	factory.start()
}

func (f *Factory) processing() {
	for pool := range f.processingPool {
		fmt.Println("###")
		psc := f.Import(pool.productName)
		rc := f.importPool.Get()
		l, err := Lock(rc, pool.productName, pool.productName, 30)

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("===", l, err)
		if l {
			psc.Unsubscribe(pool.productName)
			//psc.Close()
			pd := pool.method(pool.materials...)
			fmt.Println("^^^", pd)
			pubcli := f.importPool.Get()
			f.importPool.Get().Do("PUBLISH", pool.productName, pd)
			pubcli.Close()
			f.mu.Lock()
			for i := 0; i < len(f.products[pool.productName].forklifts); i++ {
				f.products[pool.productName].forklifts[i] <- pd
			}
			delete(f.products, pool.productName)
			f.mu.Unlock()
		}

		err = Unlock(rc, pool.productName, pool.productName)
		rc.Close()
		psc.Conn.Close()
		fmt.Println("@@@", err)
	}
}

func (f *Factory) Import(channel interface{}) redis.PubSubConn {
	rc := f.importPool.Get()
	psc := redis.PubSubConn{Conn: rc}

	psc.Subscribe(channel)
	go func() {
		for {
			switch n := psc.Receive().(type) {
			case redis.Message:
				f.mu.Lock()
				for i := 0; i < len(f.products[n.Channel].forklifts); i++ {
					f.products[n.Channel].forklifts[i] <- string(n.Data)
				}
				delete(f.products, n.Channel)
				f.mu.Unlock()
				psc.Unsubscribe(channel)
				return
			default:
				return
			}
		}
	}()

	return psc
}

func (f *Factory) start() {
	for i := 0; i < f.maxPool; i++ {
		go f.processing()
	}
}

func NewProduct(name string, method ProcessingMethod, materials ...interface{}) Forklift {
	factory.mu.Lock()
	fl := make(Forklift)
	_, ok := factory.products[name]
	fls := factory.products[name].forklifts
	fls = append(fls, fl)
	factory.products[name] = Product{
		name:      name,
		forklifts: fls,
	}
	factory.mu.Unlock()

	if !ok {
		fmt.Println("$$$")
		factory.processingPool <- ProcessingPool{
			productName: name,
			method:      method,
			materials:   materials,
		}
		fmt.Println("&&&")
	}

	return fl
}

func (f Forklift) Get() interface{} {
	defer close(f)

	return <-f
}
