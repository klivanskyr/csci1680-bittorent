package backend

import (
    "github.com/zeebo/bencode"
    "fmt"
)

func UnmarshalTorrent(data []byte) (map[string]interface{}, error) {
	var decoded map[string]interface{}
	err := bencode.DecodeBytes(data, &decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal torrent: %v", err)
	}
	return decoded, nil
}