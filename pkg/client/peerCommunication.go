package client

import (
	"bittorrent/pkg/torrent"
	"bittorrent/pkg/trackingserver"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
)

type Message byte 

const (
	Choke         Message = 0
	Unchoke       Message = 1
	Interested    Message = 2
	NotInterested Message = 3
	Request       Message = 6
	Piece         Message = 7

	BlockSize = 16384 // Standard block size (16KB)
)

func sendHandshakeToSeeder(conn net.Conn, infoHash []byte, peerId string) error {
	// Send the handshake message to the peer
	handshake := make([]byte, 68)
	handshake[0] = 19
	copy(handshake[1:20], []byte("BitTorrent protocol"))
	copy(handshake[28:48], infoHash)
	copy(handshake[48:68], []byte(peerId))

	_, err := conn.Write(handshake)
	if err != nil {
		return errors.New("failed to send handshake message")
	}

	return nil
}

func receiveHandshakeFromSeeder(conn net.Conn, infoHash []byte) error {
	// Receive the handshake message from the peer
	handshake := make([]byte, 68)
	_, err := conn.Read(handshake)
	if err != nil {
		return errors.New("failed to receive handshake message")
	}

	// Check the info hash
	if !bytes.Equal(handshake[28:48], infoHash) {
		return errors.New("info hash mismatch")
	}

	return nil
}

func sendInterested(conn net.Conn) error {
	message := []byte{0, 0, 0, 1, byte(Interested)}
	_, err := conn.Write(message)
	if err != nil {
		return errors.New("failed to send interested message")
	}

	return nil
}

func receiveUnchoke(conn net.Conn) error {
	message := make([]byte, 5)
	_, err := conn.Read(message)
	if err != nil {
		return errors.New("failed to receive unchoke message")
	}

	if Message(message[4]) != Unchoke {
		return errors.New("unexpected message type")
	}

	return nil
}

func sendRequest(conn net.Conn, index uint32, begin uint32, length uint32) error {
	// The request message is 17 bytes long:
    // 4 bytes (length prefix) + 1 byte (message ID) + 4 bytes (piece index) +
    // 4 bytes (begin) + 4 bytes (block length).

	message := make([]byte, 17)

	binary.BigEndian.PutUint32(message[0:4], 13)
	message[4] = byte(Request)
	binary.BigEndian.PutUint32(message[5:9], index)
	binary.BigEndian.PutUint32(message[9:13], begin)
	binary.BigEndian.PutUint32(message[13:17], length)

	_, err := conn.Write(message)
	if err != nil {
		return errors.New("failed to send request message")
	}

	return nil
}

func receivePiece(conn net.Conn) ([]byte, error) {
	lengthBuffer := make([]byte, 4)
	_, err := conn.Read(lengthBuffer)
	if err != nil {
		return nil, errors.New("failed to receive piece length")
	}

	length := binary.BigEndian.Uint32(lengthBuffer)
	if length == 0 {
		return nil, errors.New("zero length piece")
	}	

	message := make([]byte, length+1) // +1 for the message ID
	_, err = conn.Read(message)
	if err != nil {
		return nil, errors.New("failed to receive piece message")
	}

	return message, nil
}

func parsePiece(message []byte) (uint32, uint32, []byte) {
	index := binary.BigEndian.Uint32(message[0:4])
	begin := binary.BigEndian.Uint32(message[4:8])
	block := message[8:]

	fmt.Println("Received piece:", index, begin, block)
	return index, begin, block
}	

func downloadFromSeeder(peer trackingserver.Peer, infoHash []byte, totalPieces uint32) error {
	// Connect to the seeder
	fmt.Println("\n\n\n\nConnecting to seeder...")
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", peer.IP, peer.Port))
	if err != nil { return err }
	defer conn.Close()

	// Send the handshake message
	fmt.Println("Sending handshake to seeder...")
	err = sendHandshakeToSeeder(conn, infoHash, peer.PeerID)
	if err != nil { return err }

	// Receive the handshake message
	fmt.Println("Receiving handshake from seeder...")
	err = receiveHandshakeFromSeeder(conn, infoHash)
	if err != nil { return err }

	// Send the interested message
	fmt.Println("Sending interested message...")
	err = sendInterested(conn)
	if err != nil { return err }

	// Receive the unchoke message	
	fmt.Println("Receiving unchoke message...")
	err = receiveUnchoke(conn)
	if err != nil { return err }

	// Create file to save the downloaded data
	file, err := os.Create("downloaded_file")
	if err != nil { return err }
	defer file.Close()
	fmt.Println("File created")

	// Keep requesting pieces until the download is complete
	for pieceIndex := uint32(0); pieceIndex < totalPieces; pieceIndex++ {
		for begin := uint32(0); ; begin += BlockSize {
			err = sendRequest(conn, pieceIndex, begin, BlockSize)
			if err != nil { return err }

			message, err := receivePiece(conn)
			if err != nil { return err }
			fmt.Println("Received piece message:", message)
			
			index, offset, block := parsePiece(message)
			err = saveBlock(file, index, offset, block)
			if err != nil { return err }

			// If the block is less than BlockSize, it’s the last block of the piece
			if len(block) < BlockSize {
				break
			}
		}
	}

	fmt.Println("Download complete")
	return nil
}

func saveBlock(file *os.File, index uint32, begin uint32, block []byte) error {
	// Save the block to disk
	offset := int64(index*16384 + begin)
	_, err := file.WriteAt(block, offset)
	if err != nil {
		return errors.New("failed to save block to disk")
	}

	fmt.Println("Saved block to disk:", index, begin, len(block))

	return nil
}

func DownloadFromSeeders(peers []trackingserver.Peer, torrent torrent.Torrent, totalPieces uint32) error {
	// If there are multiple peers, download from the first one for simplicity
	infoHash, err := torrent.HashInfo()
	if err != nil {
		return fmt.Errorf("error hashing info dictionary: %v", err)
	}

	return downloadFromSeeder(peers[0], infoHash, totalPieces)
}
