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

const PIECE_SIZE = 16 * 1024 // 16 KB

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
	// Reset the file offset to the beginning
	file.Seek(0, 0)

	// Prepare to read the file in pieces
	pieces := make([]byte, 0)
	buf := make([]byte, PIECE_SIZE)

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}

		// Compute the SHA-1 hash of the current piece
		hasher := sha1.New()
		hasher.Write(buf[:n])
		pieceHash := hasher.Sum(nil)

		// Append the piece hash to the pieces slice
		pieces = append(pieces, pieceHash...)
	}

	// Normalize the path and extract the file name, Path wasnt working for some reason
	normalizedPath := strings.ReplaceAll(filePath, "\\", "/")
	lastSlashIndex := strings.LastIndex(normalizedPath, "/")
	fileName := normalizedPath[lastSlashIndex+1:]

	// Define the torrent metadata
	torrent := Torrent{
		Announce: TrackerAddr, 
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

	ipv6, err := getGlobalIPv6()
	if err != nil {
		return nil, fmt.Errorf("failed to get IPv6 address: %v", err)
	}

	// Debug IPv6
	fmt.Println("IPv6: ", ipv6)

	err = seederStack.AddSeeder(Seeder{
		addr:              &net.TCPAddr{IP: net.ParseIP(ipv6), Port: seederPort},
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

func getGlobalIPv6() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		// Skip down or loopback interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				ip := ipNet.IP
				// Check if it's a global IPv6 address (not link-local, not loopback)
				if ip.To4() == nil && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
					return ip.String(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("no global IPv6 address found")
}
