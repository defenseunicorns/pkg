// Package other is a cool package
package other

import (
	"crypto/tls"
	"fmt"
)

func func1() {
	fmt.Println("hello world")
	tlsConfig := &tls.Config{}
	fmt.Printf("%v", tlsConfig)
}
