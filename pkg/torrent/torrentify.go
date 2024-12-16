package torrent

import (
	"bytes"
	"crypto/sha1"
	"io"
	"os"
	"strings"

	"github.com/zeebo/bencode"
)

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

const PIECE_SIZE = 512 * 1024 // 512 KB

func CreateTorrentFile(seederStack *SeederStack, filePath string) ([]byte, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Calculate the SHA-1 hash of the file
	hasher := sha1.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return nil, err
	}
	pieces := hasher.Sum(nil) // This is a 20-byte hash

	// Normalize the path and extract the file name, Path wasnt working for some reason
	normalizedPath := strings.ReplaceAll(filePath, "\\", "/")
	lastSlashIndex := strings.LastIndex(normalizedPath, "/")
	fileName := normalizedPath[lastSlashIndex+1:]

	// Define the torrent metadata
	torrent := Torrent{
		Announce: "1.2.3.4:5678", // This should be the URL of the tracker server
		Info: TorrentInfo{
			Name:        fileName,
			Length:      fileInfo.Size(),
			PieceLength: PIECE_SIZE, // 16 KB piece size
			Pieces:      pieces,     // Store the raw binary hash
		},
	}

	// Encode the torrent metadata to bencode format in-memory
	var buffer bytes.Buffer
	err = bencode.NewEncoder(&buffer).Encode(torrent)
	if err != nil {
		return nil, err
	}

	// Adds to seeder stack and sends POST request to tracker
	seederStack.AddSeeder(Seeder{
		infoHash:          []byte(hasher.Sum(nil)), //TEMP 
		peerID:            []byte("12345678901234567890"), //TEMP
		filepath:          filePath,
		connectedLeechers: []Leecher{},
	}) 

	return buffer.Bytes(), nil
}
