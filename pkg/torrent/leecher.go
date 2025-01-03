package torrent

import (
	"bittorrent/pkg/trackingserver"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

func DownloadFromSeeders(peers []trackingserver.Peer, torrent Torrent, totalPieces uint32) ([]byte, error) {
	// Make a bitfield to track the pieces that we have
	bitfield := make([]byte, (totalPieces+7)/8)

	// Iterate through the list of peers, downloading as many pieces from each and moving on if one fails
	for _, peer := range peers {
		downloadedData, err := downloadFromSeeder(peer, torrent, bitfield)
		if err != nil {
			continue
		} else {
			return downloadedData, nil
		}
	}

	return nil, fmt.Errorf("failed to download from all seeders")
}

func downloadFromSeeder(peer trackingserver.Peer, torrent Torrent, bitfield []byte) ([]byte, error) {
	// Connect to the peer
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", peer.IP, peer.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to seeder")
	// Send the handshake message
	err = sendHandshakeToSeeder(conn, torrent, peer.PeerID)
	if err != nil {
		return nil, err
	}

	fmt.Println("Handshake sent")
	// Receive the handshake message
	err = receiveHandshakeFromSeeder(conn, torrent)
	if err != nil {
		return nil, err
	}

	fmt.Println("Handshake received")
	// Start downloading pieces
	totalPieces := len(torrent.Info.Pieces) / 20 // Each piece hash is 20 bytes
	downloadedData := make([]byte, 0)

	log.Println("Total pieces: ", totalPieces)
	for pieceIndex := uint32(0); pieceIndex < uint32(totalPieces); pieceIndex++ {
		// Check if we already have the piece
		if hasPiece(bitfield, pieceIndex) {
			log.Println("Already have piece: ", pieceIndex+1, "/", totalPieces)
			continue
		}

		// Send a request message for the piece
		err = sendRequest(conn, pieceIndex)
		if err != nil {
			return nil, err
		}

		// Receive the piece data
		pieceData, pieceIndexR, err := receivePiece(conn)
		if err != nil {
			return nil, err
		} else if pieceIndexR != pieceIndex {
			log.Println("received piece index %d, expected %d", pieceIndexR, pieceIndex)
			return nil, fmt.Errorf("received piece index %d, expected %d", pieceIndexR, pieceIndex)
		}

		log.Println("Piece received: ", pieceIndex, "/", totalPieces)
		//log.Println("Piece data: ", string(pieceData))

		log.Println("Piece data length: ", len(pieceData))

		// Validate the piece data
		valid, err := validatePiece(torrent, pieceIndex, pieceData)
		if err != nil || !valid {
			log.Println("Piece failed validation")
			return nil, fmt.Errorf("piece %d failed validation", pieceIndex)
		}

		log.Println("Piece validated: ", pieceIndex+1, "/", totalPieces)

		// Update the bitfield
		setPiece(bitfield, pieceIndex)

		// Append the piece data
		downloadedData = append(downloadedData, pieceData...)

		log.Println("Pieces downloaded: ", pieceIndex+1, "/", totalPieces)
	}

	// Write the downloaded data to disk
	// err = os.WriteFile("downloaded_file", downloadedData, 0644)
	// if err != nil {
	// 	return err
	// }

	return downloadedData, nil
}

func sendHandshakeToSeeder(conn net.Conn, torrent Torrent, pi string) error {
	// Create the handshake message
	infoHash, err := torrent.HashInfo()
	if err != nil {
		return fmt.Errorf("failed to hash info: %v", err)
	}
	var peerIDBytes [20]byte
	copy(peerIDBytes[:], pi)

	handshake := HandshakeMessage{
		Pstr:     "BitTorrent protocol",
		InfoHash: *(*[20]byte)(infoHash),
		PeerID:   peerIDBytes,
	}

	handshakeBytes, err := handshake.Marshal()

	// Send the handshake message
	_, err = conn.Write(handshakeBytes)
	if err != nil {
		return fmt.Errorf("failed to send handshake: %v", err)
	}

	return nil
}

func receiveHandshakeFromSeeder(conn net.Conn, torrent Torrent) error {
	// Receive the handshake message
	buf := make([]byte, 68)
	_, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read handshake: %v", err)
	}

	// Unmarshal the handshake message
	handshake, err := UnmarshalHandshake(buf)
	if err != nil {
		return fmt.Errorf("failed to unmarshal handshake: %v", err)
	}

	// Check the info hash
	infoHash, err := torrent.HashInfo()
	if err != nil {
		return fmt.Errorf("failed to hash info: %v", err)
	}
	if !bytes.Equal(handshake.InfoHash[:], infoHash) {
		return fmt.Errorf("info hash mismatch")
	}

	return nil
}

func sendRequest(conn net.Conn, pieceIndex uint32) error {
	fmt.Println("PieceIndex:", pieceIndex)
	// Create the request message
	message := Message{
		Length:  13,
		ID:      6,
		Payload: make([]byte, 12),
	}
	// Set the piece index, begin, and length
	pieceIndexBytes := uint32ToBytes(pieceIndex)
	copy(message.Payload[0:4], pieceIndexBytes)
	copy(message.Payload[4:8], uint32ToBytes(0))           // begin
	copy(message.Payload[8:12], uint32ToBytes(PIECE_SIZE)) // length

	// Marshal the message
	fmt.Println("About to send request", message)
	msgBytes, err := message.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal request message: %v", err)
	}

	// Send the message
	_, err = conn.Write(msgBytes)
	if err != nil {
		return fmt.Errorf("failed to send request message: %v", err)
	}
	return nil
}

func receivePiece(conn net.Conn) ([]byte, uint32, error) {
	// Read the message
	buf := make([]byte, 4)
	_, err := conn.Read(buf)

	length := binary.BigEndian.Uint32(buf)
	fmt.Println("Received length:", length)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read message length: %v", err)
	}

	buf2 := make([]byte, length)
	_, err = conn.Read(buf2)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read message: %v", err)
	}

	buf = append(buf, buf2...)

	// Unmarshal the message
	message, err := UnmarshalMessage(buf)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal message: %v", err)
	}

	fmt.Println("Received message:", message)
	fmt.Println("Received message bytes:", string(message.Payload))

	// Check the message ID
	if message.ID != 7 {
		return nil, 0, fmt.Errorf("unexpected message ID: %d", message.ID)
	}

	// Get the piece index
	pieceIndex := uint32(binary.BigEndian.Uint32(message.Payload[0:4]))

	return message.Payload[4:], pieceIndex, nil
}

func validatePiece(torrent Torrent, pieceIndex uint32, pieceData []byte) (bool, error) {
	// Calculate the SHA-1 hash of the piece data
	hash := sha1.Sum(pieceData)

	fmt.Println("Hash:", hash)
	fmt.Println("Expected Hash:", torrent.Info.Pieces[pieceIndex*20:(pieceIndex+1)*20])

	// Get the expected hash from the torrent metadata
	expectedHash := torrent.Info.Pieces[pieceIndex*20 : (pieceIndex+1)*20]

	if !bytes.Equal(hash[:], expectedHash) {
		return false, nil
	}

	return true, nil
}

func setPiece(bitfield []byte, index uint32) {
	byteIndex := index / 8
	offset := index % 8
	bitfield[byteIndex] |= 1 << (7 - offset)
}

func uint32ToBytes(n uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, n)
	return b
}

func hasPiece(bitfield []byte, index uint32) bool {
	byteIndex := index / 8
	offset := index % 8
	return bitfield[byteIndex]&(1<<(7-offset)) != 0
}
