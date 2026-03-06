// Command zoomosc-controller is a Robo-Stream plugin that controls the Zoom
// desktop client via ZoomOSC (https://www.liminalet.com/zoomosc).
//
// ZoomOSC must be installed alongside Zoom and configured to send OSC feedback
// to this plugin's listen port (default 1234).
//
// Usage:
//
//	zoomosc-controller             (normal operation — server spawns this)
//	zoomosc-controller --probe     (print PLUGIN_READY and exit, used by installer)
//	zoomosc-controller --buttons   (print default buttons as JSON and exit)
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/robomon1/robo-stream/sdk"
)

func main() {
	probe   := flag.Bool("probe",   false, "Print plugin info and exit (used by the server installer)")
	buttons := flag.Bool("buttons", false, "Print default buttons as JSON and exit (used by the server on startup)")
	flag.Parse()

	ctrl := newZoomOSCController()

	if *probe {
		// Fast probe mode: print PLUGIN_READY and exit immediately.
		// The server calls this during installation to discover the plugin ID.
		info := ctrl.Info()
		fmt.Printf("PLUGIN_READY {\"id\":%q,\"port\":0,\"version\":%q}\n", info.ID, info.Version)
		os.Exit(0)
	}

	if *buttons {
		// Buttons mode: print the default button library seed as JSON and exit.
		// The server calls this when the plugin first loads so it can populate
		// the button library without needing the HTTP server or ZoomOSC running.
		if err := json.NewEncoder(os.Stdout).Encode(ctrl.DefaultButtons()); err != nil {
			fmt.Fprintln(os.Stderr, "error encoding buttons:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Normal operation: start the plugin HTTP server and block until SIGTERM.
	sdk.RunPlugin(ctrl)
}
