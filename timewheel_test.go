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
	//err := tw.AddTask(2*time.Second, "2-1", print("2-1"), -1, true)
	//as.Nil(err)
	err := tw.AddTask(3*time.Second, "1", print("3-1"), -1, true)
	as.Nil(err)
	err = tw.AddTask(3*time.Second, "2", print("3-2"), -1, true)
	as.Nil(err)
	err = tw.AddTask(3*time.Second, "3", print("3-3"), -1, true)
	as.Nil(err)
	err = tw.AddTask(time.Second, "4", print("4"), 4, true)
	as.Nil(err)
	<-stopCh
}

func TestAddTask(t *testing.T) {
	as := assert.New(t)
	tw := New(DefaultOptions)
	tw.Start()
	var testKey = "testKey"
	err := tw.AddTask(time.Second, testKey, print(1), -1, false)
	as.Nil(err)
	err = tw.AddTask(time.Second, testKey, print(1), -1, false)
	as.NotNil(err)
	tw.Stop()
}

func print(i interface{}) func() {
	return func() {
		fmt.Println(i, time.Now().Second())
	}
}
