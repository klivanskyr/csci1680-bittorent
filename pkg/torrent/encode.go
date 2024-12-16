package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/zeebo/bencode"
)

const PIECE_SIZE = 512 * 1024 // 512 KB

func CreateTorrentFile(seederStack *SeederStack, filePath string, peerID string) ([]byte, error) {
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
		Announce: "http://127.0.0.1:8080/announce", // This should be the URL of the tracker server
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
	seederStack.mtx.Lock()
	seederPort := seederStack.port
	seederStack.mtx.Unlock()

	hashInfo, err := torrent.HashInfo()
	if err != nil {
		return nil, err
	}

	fmt.Println("Hash Info: ", hashInfo)

	err = seederStack.AddSeeder(Seeder{
		addr:              &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: seederPort}, //Right now just local host
		infoHash:          hashInfo,
		peerID:            []byte(peerID),
		filepath:          filePath,
		connectedLeechers: []Leecher{},
	})
	if err != nil {
		/* This error means that if we couldn't upload to the tracker server,
		we should not give you the torrent file */
		fmt.Println("Error adding seeder to stack: ", err)
		return nil, err
	}

	return buffer.Bytes(), nil
}
