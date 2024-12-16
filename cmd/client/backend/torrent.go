package backend

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/zeebo/bencode"
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
func HashInfo(torrentPath string) ([]byte, error) {
	mi, err := metainfo.LoadFromFile(torrentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load torrent file: %v", err)
	}

	info := mi.HashInfoBytes().Bytes()

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
func SendTrackerRequest(torrent interface{}, infoHash []byte) (map[string]interface{}, error) {
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

func URLEncodeBytes(data []byte) string {
	encoded := ""
	for _, b := range data {
		encoded += fmt.Sprintf("%%%02x", b)
	}
	return encoded
}

func sendHTTPTrackerRequest(peerId string, announce string, infoHash []byte) (map[string]interface{}, error) {
	fmt.Println("Sending HTTP tracker request... 1")

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

	// Decode the Bencoded response
	var trackerResponse map[string]interface{}
	err = bencode.NewDecoder(bytes.NewReader(body)).Decode(&trackerResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding tracker response: %v", err)
	}

	// parse the compact peer list
	peers, ok := trackerResponse["peers"]
	if !ok {
		return nil, fmt.Errorf("peers not found in tracker response")	
	}

	peerList, err := parseCompactPeers([]byte(peers.(string)))
	if err != nil {
		return nil, fmt.Errorf("error parsing compact peers: %v", err)
	}

	fmt.Println("Peers:", peerList)
	
	return trackerResponse, nil
}

func parseCompactPeers(peers []byte) ([]string, error) {
	const peerSize = 6 // 4 bytes for IP + 2 bytes for port

	if len(peers)%peerSize != 0 {
		return nil, fmt.Errorf("invalid peers length: not a multiple of 6")
	}

	var peerList []string

	for i := 0; i < len(peers); i += peerSize {
		ip := net.IP(peers[i : i+4])
		port := binary.BigEndian.Uint16(peers[i+4 : i+6])
		peerList = append(peerList, fmt.Sprintf("%s:%d", ip, port))
	}

	return peerList, nil
}