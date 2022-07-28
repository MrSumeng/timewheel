/*
Copyright 2022 The KubePort Authors.
*/

package timewheel

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	as := assert.New(t)
	stopCh := make(chan struct{})
	tw := New(DefaultOptions)
	as.NotNil(tw)
	tw.Start()
	_, err := tw.Add(print("1"), WithTimes(2))
	as.Nil(err)
	<-stopCh
}

func TestAddTask(t *testing.T) {
	as := assert.New(t)
	tw := New(DefaultOptions)
	tw.Start()
	var testKey = "testKey"
	_, err := tw.Add(print("1"), WithKey(testKey))
	as.Nil(err)
	_, err = tw.Add(print("1"), WithKey(testKey))
	as.NotNil(err, "duplicate task key")
	tw.Stop()
}

func print(i interface{}) func() {
	return func() {
		fmt.Println(i, time.Now().Second())
	}
}
