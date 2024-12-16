package trackingserver

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/zeebo/bencode"
)

// Announce is the message sent by the client to the tracking server to announce its presence.
type Announce struct {
	InfoHash string `bencode:"info_hash: hex"` // The info_hash of the file the client is downloading
	PeerID   string `bencode:"peer_id: hex"`   // The peer_id of the client
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
	PeerID       string    `bencode:"peer_id: hex"`       // The peer_id of the peer
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

/// Listen is a function that listens for Announce messages from clients.
func (tracker *Tracker) Listen() {
	http.HandleFunc("/announce", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleAnnounceGET(w, r, tracker)
		case http.MethodPost:
			handleAnnouncePOST(w, r, tracker)
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleAnnouncePOST handles POST requests to the /announce endpoint.
// Used when seeder wants to register itself with the tracker.
func handleAnnouncePOST(w http.ResponseWriter, r *http.Request, tracker *Tracker) {
	// Parse the bencoded request body
	var announce Announce
	err := bencode.NewDecoder(r.Body).Decode(&announce)
	if err != nil {
		http.Error(w, "Invalid bencoded payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if announce.InfoHash == "" || announce.PeerID == "" || announce.Port == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Add the seeder to the tracker's list
	tracker.mtx.Lock()
	defer tracker.mtx.Unlock()
	tracker.peers[announce.InfoHash] = append(tracker.peers[announce.InfoHash], Peer{
		PeerID:       announce.PeerID,
		Seeder:       true,
		IP:           announce.IP,
		Port:         announce.Port,
		LastAnnounce: time.Now(),
	})
	tracker.mtx.Unlock()

	// Respond with a bencoded confirmation message
	response := map[string]string{"status": "Seeder added successfully"}
	w.Header().Set("Content-Type", "application/x-bittorrent")
	err = bencode.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Error encoding response:", err)
	}
}

// handleAnnounceGET handles GET requests to the /announce endpoint.
func handleAnnounceGET(w http.ResponseWriter, r *http.Request, tracker *Tracker) {
	// Parse query parameters
	infoHash := r.URL.Query().Get("info_hash")
	peerID := r.URL.Query().Get("peer_id")
	ip := strings.Split(r.RemoteAddr, ":")[0]
	port := r.URL.Query().Get("port")
	event := r.URL.Query().Get("event")

	// Validate required fields
	if infoHash == "" || peerID == "" || port == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Convert port to integer
	portInt := 0
	_, err := fmt.Sscanf(port, "%d", &portInt)
	if err != nil {
		http.Error(w, "Invalid port number", http.StatusBadRequest)
		return
	}

	// Convert event to integer (optional)
	eventInt := STARTED
	if event != "" {
		_, err := fmt.Sscanf(event, "%d", &eventInt)
		if err != nil {
			http.Error(w, "Invalid event", http.StatusBadRequest)
			return
		}
	}

	// Create the Announce struct
	announce := Announce{
		InfoHash: infoHash,
		PeerID:   peerID,
		IP:       ip,
		Port:     portInt,
		Event:    eventInt,
	}

	// Handle the Announce message
	handleAnnounce(w, tracker, &announce)
}

// handleAnnounce is a function that handles an Announce message from a client.
func handleAnnounce(w http.ResponseWriter, tracker *Tracker, announce *Announce) {
	// Get the list of peers for the info_hash
	tracker.mtx.Lock()
	defer tracker.mtx.Unlock()

	peers := tracker.peers[announce.InfoHash]
	seeders := []Peer{}

	fmt.Print("Received Announce Message from ", announce.IP, ":", announce.Port, "\n")

	switch announce.Event {
	case STARTED:
		// Add the peer to the list of peers
		peers = append(peers, Peer{announce.PeerID, false, announce.IP, announce.Port, time.Now()})
		// Return a list of all of the seeders
		for _, peer := range peers {
			if peer.Seeder {
				seeders = append(seeders, peer)
			}
		}
	case STOPPED:
		// Remove the peer from the list of peers
		for i, peer := range peers {
			if peer.PeerID == announce.PeerID {
				peers = append(peers[:i], peers[i+1:]...)
				break
			}
		}
		// Return a list of all of the seeders
		for _, peer := range peers {
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

	// Send the list of seeders to the client
	announceResponse := AnnounceResponse{seeders}
	sendAnnounceResponse(w, &announceResponse)
}

func sendAnnounceResponse(w http.ResponseWriter, announceResponse *AnnounceResponse) {
	// Encode the announceResponse to bencode
	data, err := bencode.EncodeBytes(announceResponse)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		log.Printf("Error marshalling announce response: %v", err)
		return
	}

	// Set the response headers
	w.Header().Set("Content-Type", "application/x-bittorrent")
	w.WriteHeader(http.StatusOK)

	// Write the response data
	_, err = w.Write(data)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
