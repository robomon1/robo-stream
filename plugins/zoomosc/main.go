// Command zoomosc-controller is a Robo-Stream plugin that controls the Zoom
// desktop client via ZoomOSC (https://www.liminalet.com/zoomosc).
//
// ZoomOSC must be installed alongside Zoom and configured to send OSC feedback
// to this plugin's listen port (default 1234).
//
// Usage:
//
//	zoomosc-controller           (normal operation — server spawns this)
//	zoomosc-controller --probe   (print PLUGIN_READY and exit, used by installer)
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/robomon1/robo-stream/sdk"
)

func main() {
	probe := flag.Bool("probe", false, "Print plugin info and exit (used by the server installer)")
	flag.Parse()

	ctrl := newZoomOSCController()

	if *probe {
		// Fast probe mode: print PLUGIN_READY and exit immediately.
		// The server calls this during installation to discover the plugin ID.
		info := ctrl.Info()
		fmt.Printf("PLUGIN_READY {\"id\":%q,\"port\":0,\"version\":%q}\n", info.ID, info.Version)
		os.Exit(0)
	}

	// Normal operation: start the plugin HTTP server and block until SIGTERM.
	sdk.RunPlugin(ctrl)
}
