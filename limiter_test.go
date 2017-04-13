package mass

import (
	"fmt"
	"testing"
)

func TestLimiter(t *testing.T) {
	l := NewLimiter(3)

	for {
		l.Limit()
		fmt.Println("hello qianlnk")
	}
}
