package iidxfavlist

import (
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	srv, err := New()

	if err != nil {
		panic(err)
	}

	go srv.Run()

	time.Sleep(time.Second)

	m.Run()
}
