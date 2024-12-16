package trackingserver

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/zeebo/bencode"
)

const TIMEOUT = 2 * time.Minute

// Announce is the message sent by the client to the tracking server to announce its presence.
type Announce struct {
	InfoHash string `bencode:"info_hash"` // The info_hash of the file the client is downloading
	PeerID   string `bencode:"peer_id"`   // The peer_id of the client
	IP       string `bencode:"ip"`        // The IP address of the client
	Port     int    `bencode:"port"`      // The port the client is listening on
	Event    int    `bencode:"event"`     // The event of the announce message
}

const (
	STARTED   = 0
	STOPPED   = 1
	COMPLETED = 2
)

// AnnounceResponse is the message sent by the tracking server to the client in response to an Announce message.
type AnnounceResponse struct {
	Peers []Peer `bencode:"peers"` // A list of peers that have the file the client is downloading
}

// Peer is a struct that represents a peer that has the file the client is downloading.
type Peer struct {
	PeerID       string    `bencode:"peer_id"`       // The peer_id of the peer
	Seeder       bool      `bencode:"seeder"`        // Whether the peer has downloaded the file
	IP           string    `bencode:"ip"`            // The IP address of the peer
	Port         int       `bencode:"port"`          // The port the peer is listening on
	LastAnnounce time.Time `bencode:"last_announce"` // The time of the last announce message from the peer
}

// Tracker is a struct that represents a tracking server, keeping a map of info_hashes to a list of peers.
type Tracker struct {
	mtx   sync.Mutex            // A mutex to protect the peers map
	peers map[string][]Peer     // A map of info_hashes to a list of peers
}

// NewTracker is a function that creates a new tracking server.
func NewTracker() *Tracker {
	return &Tracker{
		peers: make(map[string][]Peer),
	}
}

// Listen is a function that listens for Announce messages from clients.
func (tracker *Tracker) Listen() {
	http.HandleFunc("/announce", func(w http.ResponseWriter, r *http.Request) {
		// Parse the Announce message from the request body
		var announce Announce
		err := bencode.NewDecoder(r.Body).Decode(&announce)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Handle the Announce message
		handleAnnounce(tracker, &announce)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleAnnounce is a function that handles an Announce message from a client.
func handleAnnounce(tracker *Tracker, announce *Announce) {
	// Get the list of peers for the info_hash
	tracker.mtx.Lock()
	peers := tracker.peers[announce.InfoHash]
	seeders := []Peer{}

	switch announce.Event {
	case STARTED:
		// Add the peer to the list of peers
		peers = append(peers, Peer{announce.PeerID, false, announce.IP, announce.Port, time.Now()}) // This allows repeats but doesn't matter
		// Return a list of all of the seeders
		for i, peer := range peers {
			if peer.LastAnnounce.Before(time.Now().Add(-TIMEOUT)) {
				// Remove the peer from the list of peers
				peers = append(peers[:i], peers[i+1:]...)
				continue
			}
			if peer.Seeder {
				seeders = append(seeders, peer)
			}
		}
	case STOPPED:
		// Remove the peer from the list of peers
		for i, peer := range peers {
			if peer.LastAnnounce.Before(time.Now().Add(-TIMEOUT)) {
				// Remove the peer from the list of peers
				peers = append(peers[:i], peers[i+1:]...)
				continue
			}
			if peer.PeerID == announce.PeerID {
				peers = append(peers[:i], peers[i+1:]...)
				break
			}
		}
		// Return a list of all of the seeders
		for i, peer := range peers {
			if peer.LastAnnounce.Before(time.Now().Add(-TIMEOUT)) {
				// Remove the peer from the list of peers
				peers = append(peers[:i], peers[i+1:]...)
				continue
			}
			if peer.Seeder {
				seeders = append(seeders, peer)
			}
		}
	case COMPLETED:
		// Mark the peer as a seeder
		for i, peer := range peers {
			if peer.PeerID == announce.PeerID {
				peers[i].Seeder = true
				break
			}
		}
	}
	tracker.peers[announce.InfoHash] = peers
	tracker.mtx.Unlock()

	// Send the list of seeders to the client
	announceResponse := AnnounceResponse{seeders}
	sendAnnounceResponse(&announceResponse, announce.IP, announce.Port)
}

func sendAnnounceResponse(announceResponse *AnnounceResponse, ip string, port int) {
	// Send the list of seeders to the client
	data, err := bencode.EncodeBytes(announceResponse)
	if err != nil {
		log.Printf("Error marshalling announce response: %v", err)
		return
	}

	url := fmt.Sprintf("https://%s:%d/announce", ip, port)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-bittorrent")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending HTTP request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-OK response: %s", resp.Status)
	}
}
