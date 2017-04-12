package mass

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNewProduct(t *testing.T) {
	StartFactory("127.0.0.1:6379", 2, 10, 100)

	go func() {
		p := NewProduct("1001", howToCook, 1, "1001", "1001")
		var res string
		p.Get(&res)
		fmt.Println(res)
	}()

	go func() {
		p := NewProduct("1002", howToCook, 1, "1002", "1002")
		var res string
		p.Get(&res)
		fmt.Println(res)
	}()

	p := NewProduct("1003", howToCook, 1, "1003", "1003")
	var res string
	p.Get(&res)
	fmt.Println(res)
}

func howToCook(args ...interface{}) interface{} {
	var res string
	for _, a := range args {
		res += fmt.Sprintf("%v", a)
	}
	time.Sleep(time.Second)
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
