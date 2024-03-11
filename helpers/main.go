// Package helpers is a cool package
package helpers

import (
	"sync"
)

func func1() {
	var lock sync.Mutex

	l := lock
	l.Lock()
	l.Unlock()
}
