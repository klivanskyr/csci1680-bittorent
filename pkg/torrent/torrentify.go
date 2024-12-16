package torrent

import (
	"crypto/sha1"
	"os"
	"io"

	"github.com/jackpal/bencode-go"
)

const PIECE_SIZE = 512 * 1024 // 512 KB

func CreateTorrentFile(filePath string, torrentPath string, trackerURL string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	pieceLength := int64(PIECE_SIZE)
	buf := make([]byte, pieceLength)
	pieces := []byte{}

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		hash := sha1.Sum(buf[:n])
		pieces = append(pieces, hash[:]...)

		if n < int(pieceLength) {
			break
		}
	}

	info := map[string]interface{}{
		"name":         fileInfo.Name(),
		"length":       fileInfo.Size(),
		"piece length": pieceLength,
		"pieces":       pieces,
	}

	torrent := map[string]interface{}{
		"announce": trackerURL,
		"info":     info,
	}

	torrentFile, err := os.Create(torrentPath)
	if err != nil {
		return err
	}
	defer torrentFile.Close()

	err = bencode.Marshal(torrentFile, torrent)
	if err != nil {
		return err
	}

	return nil
}