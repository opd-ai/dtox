package main

import "log"

// doBootstrap connects to the default Tox DHT bootstrap nodes provided by
// the toxcore library. Errors are logged but do not cause failure; the DHT
// will find peers over time.
func (b *ToxBackend) doBootstrap() {
	if err := b.tox.BootstrapDefaults(); err != nil {
		log.Printf("bootstrap defaults failed: %v", err)
	} else {
		log.Println("bootstrap defaults ok")
	}
}
