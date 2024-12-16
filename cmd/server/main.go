package main

import (
	TrackingServer "bittorrent/pkg/trackingserver"
)

func main() {
	tracker := TrackingServer.NewTracker()
	tracker.Listen()
}