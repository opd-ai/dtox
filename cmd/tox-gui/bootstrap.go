package main

import "log"

// bootstrapNode describes a Tox DHT bootstrap node.
type bootstrapNode struct {
	Host      string
	Port      uint16
	PublicKey string
}

// bootstrapNodes is a hardcoded list of well-known Tox bootstrap nodes.
// These are used to join the DHT network on startup.
var bootstrapNodes = []bootstrapNode{
	{"node.tox.biribiri.org", 33445, "F404ABAA1C99A9D37D61AB54898F56793E1DEF8BD46B1038B9D822E8460FAB67"},
	{"tox.verdict.gg", 33445, "1C5293AEF2114717547B39DA8EA6F1E331E5E358B35F9B6B5F19317911C5F976"},
	{"tox.initramfs.io", 33445, "3F0A45A268367C1BEA652F258C85F4A66DA76BCAA667A49E770BCC4917AB6A25"},
	{"tox.abilinski.com", 33445, "10C00EB250C3233E343E2AEBA07115A5C28920E9C8D29492F6D00B29049EDC7E"},
	{"tox.novg.net", 33445, "D527E5847F8330D628DAB1814F0A422F6DC9D0A300E6C357634EE2DA88C35463"},
}

// doBootstrap attempts to connect to all known bootstrap nodes.
// Errors are logged but do not cause failure; the DHT will find peers over time.
func (b *ToxBackend) doBootstrap() {
	for _, node := range bootstrapNodes {
		if err := b.tox.Bootstrap(node.Host, node.Port, node.PublicKey); err != nil {
			log.Printf("bootstrap %s:%d failed: %v", node.Host, node.Port, err)
		} else {
			log.Printf("bootstrap %s:%d ok", node.Host, node.Port)
		}
	}
}
