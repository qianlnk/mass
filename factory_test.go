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
		p := NewProduct("1001", howToCook, "1001", "1001")
		fmt.Printf("----")
		fmt.Println(p.Get())
	}()

	go func() {
		p := NewProduct("1001002", howToCook, "1001002", "1001002")
		fmt.Printf("----")
		fmt.Println(p.Get())
	}()

	p := NewProduct("1001", howToCook, "1001", "1001")
	fmt.Printf("----")
	fmt.Println(p.Get())
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

func TestDelEmptyMap(t *testing.T) {
	m := make(map[string]int)
	m["test"] = 1
	delete(m, "test")
	m["test"] = 2
}
