package mass

import (
	"sync"
	"time"

	"fmt"

	"runtime"

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
	maxActive      int
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

	cpu := runtime.NumCPU()
	//runtime.GOMAXPROCS(cpu / 2)
	factory = &Factory{
		products:       make(map[string]Product),
		processingPool: make(chan ProcessingPool),
		maxActive:      cpu,
		importPool:     rp,
	}

	factory.start()
}

func (f *Factory) processing() {
	for pool := range f.processingPool {
		psc := f.Import(pool.productName)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				switch n := psc.Receive().(type) {
				case redis.Message:
					psc.Unsubscribe(pool.productName)
					f.mu.Lock()
					for i := 0; i < len(f.products[n.Channel].forklifts); i++ {
						f.products[n.Channel].forklifts[i] <- string(n.Data)
					}
					delete(f.products, n.Channel)
					f.mu.Unlock()
				case redis.Subscription:
					if n.Kind == "unsubscribe" {
						return
					}
				}
			}
		}()

		l, err := f.LockImporter(pool.productName, pool.productName, 30)
		if err != nil {
			fmt.Println(err)
		}

		if l {
			pd := pool.method(pool.materials...)
			f.UnlockImporter(pool.productName, pool.productName) //先解锁再发布，防止一些订阅了确收不到消息

			f.Export(pool.productName, pd)
		}

		wg.Wait()
		psc.Conn.Close()
	}
}

func (f *Factory) LockImporter(product string, secret string, ttl uint64) (bool, error) {
	rc := f.importPool.Get()
	defer rc.Close()

	return Lock(rc, product, secret, ttl)
}

func (f *Factory) UnlockImporter(product string, secret string) error {
	rc := f.importPool.Get()
	defer rc.Close()

	return Unlock(rc, product, secret)
}

func (f *Factory) Import(channel interface{}) redis.PubSubConn {
	rc := f.importPool.Get()
	psc := redis.PubSubConn{Conn: rc}

	psc.Subscribe(channel)

	return psc
}

func (f *Factory) Export(channel interface{}, msg interface{}) {
	rc := f.importPool.Get()
	defer rc.Close()

	rc.Do("PUBLISH", channel, msg)
}

func (f *Factory) start() {
	for i := 0; i < f.maxActive; i++ {
		go f.processing()
	}
}

func NewProduct(name string, method ProcessingMethod, materials ...interface{}) Forklift {
	factory.mu.Lock()
	fl := make(Forklift)
	_, ok := factory.products[name]
	fls := factory.products[name].forklifts
	fls = append(fls, fl)
	//delete(factory.products, name)
	factory.products[name] = Product{
		name:      name,
		forklifts: fls,
	}
	factory.mu.Unlock()

	if !ok {
		factory.processingPool <- ProcessingPool{
			productName: name,
			method:      method,
			materials:   materials,
		}
	}

	return fl
}

func (f Forklift) Get() interface{} {
	defer close(f)

	return <-f
}
