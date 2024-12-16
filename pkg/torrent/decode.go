package torrent

import (
	"bytes"
	"fmt"

	"github.com/zeebo/bencode"
)

// UnmarshalTorrent decodes Bencoded data into Go native types.
func UnmarshalTorrent(data []byte) (*Torrent, error) {
    // Decode the Bencoded data into Go native types
    var torrent Torrent
    err := bencode.NewDecoder(bytes.NewReader(data)).Decode(&torrent)
    if err != nil {
        return nil, fmt.Errorf("failed to decode torrent file: %v", err)
    }

    return &torrent, nil
}

