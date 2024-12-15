package backend

import (
	"bytes"
	"crypto/sha1"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"time"
)

// UnmarshalTorrent decodes Bencoded data into Go native types.
func UnmarshalTorrent(data []byte) (interface{}, error) {
	reader := bytes.NewReader(data)
	return decodeValue(reader)
}

func decodeValue(r *bytes.Reader) (interface{}, error) {
	ch, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch {
	case ch == 'i':
		return decodeInt(r)
	case ch >= '0' && ch <= '9':
		r.UnreadByte()
		return decodeString(r)
	case ch == 'l':
		return decodeList(r)
	case ch == 'd':
		return decodeDict(r)
	default:
		return nil, fmt.Errorf("invalid Bencode format: unexpected character %q", ch)
	}
}

// decodeInt decodes a Bencoded integer (e.g., i42e).
func decodeInt(r *bytes.Reader) (int, error) {
	var buf bytes.Buffer
	for {
		ch, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		if ch == 'e' {
			break
		}
		buf.WriteByte(ch)
	}
	num, err := strconv.Atoi(buf.String())
	if err != nil {
		return 0, errors.New("invalid integer format")
	}
	return num, nil
}

// decodeString decodes a Bencoded string (e.g., 4:spam).
func decodeString(r *bytes.Reader) (string, error) {
	var lengthBuf bytes.Buffer

	// Read the length part
	for {
		ch, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if ch == ':' {
			break
		}
		if ch < '0' || ch > '9' {
			return "", errors.New("invalid string length format")
		}
		lengthBuf.WriteByte(ch)
	}

	// Convert length to integer
	length, err := strconv.Atoi(lengthBuf.String())
	if err != nil {
		return "", errors.New("invalid string length")
	}

	// Read the specified number of bytes
	str := make([]byte, length)
	_, err = io.ReadFull(r, str)
	if err != nil {
		return "", err
	}

	return string(str), nil
}

// decodeList decodes a Bencoded list (e.g., l4:spami42ee).
func decodeList(r *bytes.Reader) ([]interface{}, error) {
	var list []interface{}
	for {
		ch, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if ch == 'e' {
			break
		}
		r.UnreadByte()
		value, err := decodeValue(r)
		if err != nil {
			return nil, err
		}
		list = append(list, value)
	}
	return list, nil
}

// decodeDict decodes a Bencoded dictionary (e.g., d3:bar4:spam3:fooi42ee).
func decodeDict(r *bytes.Reader) (map[string]interface{}, error) {
	dict := make(map[string]interface{})
	for {
		ch, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if ch == 'e' {
			break
		}
		r.UnreadByte()
		key, err := decodeString(r)
		if err != nil {
			return nil, err
		}
		value, err := decodeValue(r)
		if err != nil {
			return nil, err
		}
		dict[key] = value
	}
	return dict, nil
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
	info, ok := torrentMap["info"]
	if !ok {
		return nil, fmt.Errorf("'info' dictionary not found in torrent")
	}

	// Encode the "info" dictionary into Bencode
	fmt.Println("Calculating info hash... 3")
	// Print the type of info fields
	for key, value := range info.(map[string]interface{}) {
		fmt.Println(key, reflect.TypeOf(value))	
	}

	encodedInfo, err := Encode(info)
	if err != nil {
		return nil, fmt.Errorf("failed to Bencode 'info' dictionary: %v", err)
	}

	// Compute the SHA-1 hash of the encoded "info" dictionary
	fmt.Println("Calculating info hash... 4")
	hash := sha1.Sum(encodedInfo)
	return hash[:], nil
}

// Encode is a helper function to encode data into Bencode format.
// If you already have an Encode function, use that instead.
func Encode(data interface{}) ([]byte, error) {
	switch v := data.(type) {
	case float64: //For some reason piece length is float64 so we need to convert it to int
		return []byte(fmt.Sprintf("i%de", int(v))), nil
	case int:
		return []byte(fmt.Sprintf("i%de", v)), nil
	case string:
		return []byte(fmt.Sprintf("%d:%s", len(v), v)), nil
	case []interface{}:
		result := []byte("l")
		for _, item := range v {
			encoded, err := Encode(item)
			if err != nil {
				return nil, err
			}
			result = append(result, encoded...)
		}
		result = append(result, 'e')
		return result, nil
	case map[string]interface{}:
		result := []byte("d")
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		// Sort keys for consistent encoding
		sort.Strings(keys)
		for _, key := range keys {
			keyEncoded, err := Encode(key)
			if err != nil {
				return nil, err
			}
			valueEncoded, err := Encode(v[key])
			if err != nil {
				return nil, err
			}
			result = append(result, keyEncoded...)
			result = append(result, valueEncoded...)
		}
		result = append(result, 'e')
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported data type: %s", reflect.TypeOf(data))
	}
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