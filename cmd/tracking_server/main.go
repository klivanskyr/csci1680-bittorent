import {
	  TrackingServer
}

func main() {
	tracker := TrackingServer.NewTracker()
	tracker.Listen()
}