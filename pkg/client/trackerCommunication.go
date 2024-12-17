package client

import (
	"bittorrent/pkg/torrent"
	TrackingServer "bittorrent/pkg/trackingserver"
	"log"
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/zeebo/bencode"
)

/*
THIS FILE SHOULD REALLY BE APART OF LEACHER.GO
*/
func GeneratePeerID() string {
	const clientPrefix = "-GO0001-" // Go client

	randomBytes := make([]byte, 12)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic("Failed to generate random bytes for peer_id")
	}

	// Make peer_id hexadecimal
	return clientPrefix + fmt.Sprintf("%x", randomBytes)
}

// SendTrackerRequest sends a GET request to the tracker's announce URL.
func SendTrackerRequest(torrent torrent.Torrent, peerId string) ([]TrackingServer.Peer, error) {
	// Announce can be either UDP or HTTP
	// We are going to ignore announce list for now
	var trackerType string
	if torrent.Announce[:4] == "http" {
		trackerType = "http"
	} else if torrent.Announce[:3] == "udp" {
		trackerType = "udp"
	} else {
		return nil, fmt.Errorf("unsupported tracker protocol")
	}

	if trackerType == "http" {
		infoHash, err := torrent.HashInfo()
		if err != nil {
			return nil, fmt.Errorf("error hashing info dictionary: %v", err)
		}
		return sendHTTPTrackerRequest(peerId, torrent.Announce, infoHash)
	} else if trackerType == "udp" {
		return nil, fmt.Errorf("unsupported tracker protocol UDP")
	} else {
		return nil, fmt.Errorf("unsupported tracker protocol OTHER")
	}
}

func URLEncodeBytes(data []byte) string {
	encoded := ""
	for _, b := range data {
		encoded += fmt.Sprintf("%%%02x", b)
	}
	return encoded
}

func sendHTTPTrackerRequest(peerId string, announce string, infoHash []byte) ([]TrackingServer.Peer, error) {
	// Manually encode each byte of the info_hash
	encodedInfoHash := URLEncodeBytes(infoHash)

	// Manually construct the query parameters
	query := fmt.Sprintf(
		"info_hash=%s&peer_id=%s&port=6881&uploaded=0&downloaded=0&left=0&compact=1",
		encodedInfoHash,
		url.QueryEscape(peerId),
	)

	// Construct the full request URL
	requestURL := fmt.Sprintf("%s?%s", announce, query)

	fmt.Println("Request URL:", requestURL)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	// Set a BitTorrent-compatible User-Agent
	req.Header.Set("User-Agent", "BitTorrent/7.10.5")

	// Create an HTTP client with a timeout
	client := http.Client{Timeout: 10 * time.Second}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending tracker request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading tracker response: %v", err)
	}

	// Decode the Bencoded response
	var trackerResponse TrackingServer.AnnounceResponse
	err = bencode.NewDecoder(bytes.NewReader(body)).Decode(&trackerResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding tracker response: %v", err)
	}

	// Log the response
	log.Printf("Tracker response: %v\n", trackerResponse.Peers)

	return trackerResponse.Peers, nil
}
