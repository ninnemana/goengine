package globals

import (
	"flag"
)

var (
	ListenAddr = flag.String("http", ":8080", "http listen address")
)

func SetGlobals() {
	flag.Parse()
}
