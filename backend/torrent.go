package backend

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/zeebo/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

// UnmarshalTorrent decodes Bencoded data into Go native types.
func UnmarshalTorrent(data []byte) (interface{}, error) {
    // Decode the Bencoded data into Go native types
    var torrent interface{}
    err := bencode.NewDecoder(bytes.NewReader(data)).Decode(&torrent)
    if err != nil {
        return nil, fmt.Errorf("failed to decode torrent file: %v", err)
    }

    return torrent, nil
}

// Take in a whole torrent file and return the hashed info dictionary
func HashInfo(torrentPath string) (string, error) {
	mi, err := metainfo.LoadFromFile(torrentPath)
	if err != nil {
		return "", fmt.Errorf("failed to load torrent file: %v", err)
	}

	info := mi.HashInfoBytes().HexString()

	fmt.Println("Hashed info dictionary: ", info)
	return info, nil
}

func generatePeerID() string {
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
func SendTrackerRequest(torrent interface{}, infoHash string) ([]byte, error) {
	fmt.Println("Sending tracker request... 1")
	announce, ok := torrent.(map[string]interface{})["announce"].(string)
	if !ok {
		return nil, fmt.Errorf("announce URL not found in torrent file")
	}

	// Announce can be either UDP or HTTP
	// We are going to ignore announce list for now
	var trackerType string
	if announce[:4] == "http" {
		trackerType = "http"
	} else if announce[:3] == "udp" {
		trackerType = "udp"
	} else {
		return nil, fmt.Errorf("unsupported tracker protocol")
	}

	if trackerType == "http" {
		peerId := generatePeerID()
		return sendHTTPTrackerRequest(peerId, announce, infoHash)
	} else if trackerType == "udp" {
		return nil, fmt.Errorf("unsupported tracker protocol UDP")
	} else {
		return nil, fmt.Errorf("unsupported tracker protocol OTHER")
	}
}

func sendHTTPTrackerRequest(peerId string, announce string, infoHash string) ([]byte, error) {
	fmt.Println("Sending HTTP tracker request... 1")

	// Construct the request parameters
	params := url.Values{}
	params.Add("info_hash", infoHash)
	params.Add("peer_id", peerId)
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", "0") // Assuming download hasn't started yet
	params.Add("compact", "1")

	// Construct the full request URL
	requestURL := fmt.Sprintf("%s?%s", announce, params.Encode())

	fmt.Println("Sending HTTP tracker request... 2")
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

	return body, nil
}