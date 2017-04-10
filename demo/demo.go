package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qianlnk/mass"
	"github.com/qianlnk/redis"
)

func main() {
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

	err := redis.Get(key, &res)
	if err != nil {
		res = prefix + "qianlnk" + newRandomString(5)
		err = redis.Set(key, res, time.Second*120)
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
