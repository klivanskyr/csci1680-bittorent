package backend

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"

	// "reflect"
	"time"

	"github.com/zeebo/bencode"
)

func Dothething(data []byte) ([]byte, error) {
	torrent, err := UnmarshalTorrent(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal torrent: %v", err)
	}

	trackerResponse, err := SendTrackerRequest(torrent)
	if err != nil {
		return nil, fmt.Errorf("failed to send tracker request: %v", err)
	}

	return trackerResponse, nil
}


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

// calculateInfoHash computes the SHA-1 hash of the Bencoded "info" dictionary.
func calculateInfoHash(torrent interface{}) ([]byte, error) {
	fmt.Println("Calculating info hash... 1")

	// Ensure the torrent is a dictionary
	torrentMap, ok := torrent.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid torrent format: expected a dictionary")
	}

	fmt.Println("Calculating info hash... 2")

	// Extract the "info" dictionary
	info, ok := torrentMap["info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("'info' dictionary not found or has incorrect type in torrent")
	}

	// Fix types in the "info" dictionary
	if length, ok := info["length"].(float64); ok {
		info["length"] = int64(length)
	}

	if pieceLength, ok := info["piece length"].(float64); ok {
		info["piece length"] = int64(pieceLength)
	}

	if private, ok := info["private"].(float64); ok {
		info["private"] = int64(private)
	}

	// Initialize the writer as a bytes.Buffer
	var buf bytes.Buffer
	encoder := bencode.NewEncoder(&buf)

	fmt.Println("Calculating info hash... 3")
	// Encode the "info" dictionary into Bencode
	// Print all the keys of info and the values types
	for key, value := range info {
		fmt.Println(key, reflect.TypeOf(value))
	}

	if err := encoder.Encode(info); err != nil {
		fmt.Println("Calculating info hash... 3.1")
		return nil, fmt.Errorf("failed to encode 'info' dictionary: %v", err)
	}

	// Compute the SHA-1 hash of the encoded "info" dictionary
	fmt.Println("Calculating info hash... 4")
	hash := sha1.Sum(buf.Bytes())
	fmt.Println("Calculating info hash... 5")

	return hash[:], nil
}

func generatePeerID() string {
	const clientPrefix = "-GO0001-" // Go client

	randomBytes := make([]byte, 12)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic("Failed to generate random bytes for peer_id")
	}

	return clientPrefix + string(randomBytes)
}

// SendTrackerRequest sends a GET request to the tracker's announce URL.
func SendTrackerRequest(torrent interface{}) ([]byte, error) {
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
		return sendHTTPTrackerRequest(torrent, peerId, announce)
	} else if trackerType == "udp" {
		return nil, fmt.Errorf("unsupported tracker protocol UDP")
	} else {
		return nil, fmt.Errorf("unsupported tracker protocol OTHER")
	}
}

func sendHTTPTrackerRequest(torrent interface{}, peerId string, announce string) ([]byte, error) {
	fmt.Println("Sending HTTP tracker request... 1")

	infoHash, err := calculateInfoHash(torrent)
	fmt.Println("\n", infoHash)
	fmt.Printf("Hex info_hash: %x\n", infoHash)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate info hash: %v", err)
	}

	fmt.Println("Sending HTTP tracker request... 2")
	// URL-encode the info hash
	infoHashEncoded := url.QueryEscape(string(infoHash))

	// Construct the request parameters
	params := url.Values{}
	params.Add("info_hash", infoHashEncoded)
	params.Add("peer_id", peerId)
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", "0") // Assuming download hasn't started yet
	params.Add("compact", "1")

	requestURL := fmt.Sprintf("%s?%s", announce, params.Encode())

	fmt.Println("Sending HTTP tracker request... 3")
	fmt.Println(requestURL)

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("error sending tracker request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading tracker response: %v", err)
	}

	return body, nil
}