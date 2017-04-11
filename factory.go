package mass

import (
	"runtime"
	"sync"
	"time"

	"fmt"

	"github.com/garyburd/redigo/redis"
)

const (
	DEFAULT_LIMIT = 20000
)

//ProcessingMethod the processing method
type ProcessingMethod func(materials ...interface{}) interface{}

type Forklift chan []byte

type ProcessingPool struct {
	productName string
	method      ProcessingMethod
	materials   []interface{}
	forklift    Forklift
}

type Product struct {
	name      string
	forklifts []Forklift
}

type Factory struct {
	// mu             sync.Mutex
	// products       map[string][]Forklift
	processingPool chan ProcessingPool
	maxActive      int
	importPool     *redis.Pool
	limiter        *Limiter
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
		IdleTimeout: time.Second * 180,
	}

	cpu := runtime.NumCPU()
	runtime.GOMAXPROCS(cpu)
	factory = &Factory{
		processingPool: make(chan ProcessingPool),
		maxActive:      100,
		importPool:     rp,
		limiter:        NewLimiter(DEFAULT_LIMIT),
	}

	factory.start()
}

func (f *Factory) processing() {
	for pool := range f.processingPool {
		done := make(chan bool)
		psc := f.Import(pool.productName)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				switch n := psc.Receive().(type) {
				case redis.Message:
					psc.Unsubscribe(pool.productName)
					pool.forklift <- n.Data
					done <- true
					return
				case redis.Subscription:
					if n.Kind == "unsubscribe" {
						return
					}
				default:
				}
			}
		}()

		f.Production(pool)
		for {
			ok := false
			select {
			case <-done:
				ok = true
			case <-time.After(time.Second * 3):
				fmt.Println("timeout")
				f.Production(pool)
				ok = false
			}

			if ok {
				break
			}

		}

		wg.Wait()
		close(done)
		psc.Conn.Close()
	}
}
func (f *Factory) Production(pool ProcessingPool) error {
	l, err := f.LockImporter(pool.productName, pool.productName, 30)
	if err != nil {
		return err
	}

	if l {
		pd := pool.method(pool.materials...)
		f.Export(pool.productName, pd)

		f.UnlockImporter(pool.productName, pool.productName)
	}

	return nil
}

func (f *Factory) LockImporter(product string, secret string, ttl uint64) (bool, error) {
	rc := f.importPool.Get()
	defer rc.Close()

	return Lock(rc, "mass_lock_key:"+product, secret, ttl)
}

func (f *Factory) UnlockImporter(product string, secret string) error {
	rc := f.importPool.Get()
	defer rc.Close()

	return Unlock(rc, "mass_lock_key:"+product, secret)
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
	factory.limiter.Limit()
	fl := make(Forklift)

	factory.processingPool <- ProcessingPool{
		productName: name,
		method:      method,
		materials:   materials,
		forklift:    fl,
	}

	return fl
}

func (f Forklift) Get() interface{} {
	defer close(f)

	return string(<-f)
}
