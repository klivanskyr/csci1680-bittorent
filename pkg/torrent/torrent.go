package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"

	"github.com/zeebo/bencode"
)

// type GenericTorrent map[string]interface{} // Handles optional fields when hashing

type Torrent struct {
	Announce string      `bencode:"announce"`
	Info     TorrentInfo `bencode:"info"`
}

type TorrentInfo struct {
	Name        string `bencode:"name"`
	Length      int64  `bencode:"length"`
	PieceLength int    `bencode:"piece length"`
	Pieces      []byte `bencode:"pieces"` // Use []byte for raw binary data
}

func (torrent *Torrent) HashInfo() ([]byte, error) {
	var buf bytes.Buffer
	err := bencode.NewEncoder(&buf).Encode(torrent.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to bencode info dictionary: %v", err)
	}

	hash := sha1.Sum(buf.Bytes())
	fmt.Printf("Hashed info dictionary: %x\n", hash)

	return hash[:], nil
}