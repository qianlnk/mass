package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"sync"

	"github.com/qianlnk/mass"
)

func main() {
	mass.StartFactory("127.0.0.1:6379", 2, 10, 10000)

	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := mass.NewProduct(strconv.Itoa(i), howToCook, i)
			fmt.Println("----")
			fmt.Println(p.Get())
		}(i)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := mass.NewProduct(strconv.Itoa(i), howToCook, i)
			fmt.Println("----")
			fmt.Println(p.Get())
		}(i)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := mass.NewProduct(strconv.Itoa(i), howToCook, i)
			fmt.Println("----")
			fmt.Println(p.Get())
		}(i)
		//rand.Seed(time.Now().UnixNano())
		//time.Sleep(time.Nanosecond * time.Duration(rand.Intn(1000)))
	}
	wg.Wait()
	fmt.Println("All done")
	time.Sleep(time.Second * 3)
}

func howToCook(args ...interface{}) interface{} {
	var res string
	for _, a := range args {
		res += fmt.Sprintf("%v", a)
	}
	time.Sleep(time.Second * 1)
	return res + "qianlnk" + newRandomString(5)
}

func newRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	rs := make([]string, length)
	for start := 0; start < length; start++ {
		rs = append(rs, strconv.Itoa(rand.Intn(10)))
	}
	return strings.Join(rs, "")
}
