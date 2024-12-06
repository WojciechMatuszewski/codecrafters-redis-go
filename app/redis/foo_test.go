package redis_test

import (
	"testing"
	"time"
)

func Test(t *testing.T) {
	t.Run("test timer", func(t *testing.T) {
		timer := time.After(time.Duration(100 * int(time.Millisecond)))

		t.Log("before timer")

		<-timer

		t.Log("after timer timer")

		<-timer

		t.Log("after timer timer2")
	})
}
