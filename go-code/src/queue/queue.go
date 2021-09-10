package queue

import (
	"github.com/mborders/artifex"
)

var Dispatcher *artifex.Dispatcher

func InitDispatcher() {
	Dispatcher = artifex.NewDispatcher(10, 100)
	Dispatcher.Start()
}
