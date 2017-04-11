package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qianlnk/mass"
	"github.com/qianlnk/redikey"
)

func main() {
	flag.Parse()

	//这里实现了远程获取pprof数据的接口
	go func() {
		log.Println(http.ListenAndServe("localhost:7777", nil))
	}()

	mass.StartFactory("127.0.0.1:6379", 2, 100, 1000)
	var wg sync.WaitGroup
	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := mass.NewProduct(strconv.Itoa(i), howToProcessing, i)
			fmt.Println(p.Get())
		}(i)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := mass.NewProduct(strconv.Itoa(i), howToProcessing, i)
			fmt.Println(p.Get())
		}(i)
		// wg.Add(1)
		// go func(i int) {
		// 	defer wg.Done()
		// 	p := mass.NewProduct(strconv.Itoa(i), howToCook, i)
		// 	fmt.Println(p.Get())
		// }(i)
		//rand.Seed(time.Now().UnixNano())
		//time.Sleep(time.Nanosecond * time.Duration(rand.Intn(1000)))
	}

	wg.Wait()
	fmt.Println("All done")
	time.Sleep(time.Second * 3)
}

func howToProcessing(args ...interface{}) interface{} {
	var res, key, prefix string

	key = "mass_key:"
	for _, a := range args {
		prefix += fmt.Sprintf("%v", a)
	}

	key += prefix

	err := redikey.Get(key, &res)
	if err != nil {
		res = prefix + "qianlnk" + newRandomString(5)
		err = redikey.Set(key, res, time.Second*120)
		if err != nil {
			fmt.Println(err)
		}
	}

	//time.Sleep(time.Second * 1)
	return res
}

func newRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	rs := make([]string, length)
	for start := 0; start < length; start++ {
		rs = append(rs, strconv.Itoa(rand.Intn(10)))
	}
	return strings.Join(rs, "")
}
