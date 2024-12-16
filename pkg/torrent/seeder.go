package torrent

import (
	"bytes"
	"log"
	"net"
	"os"
)

// All the seeder needs to do is respond to requests for specific pieces of a file

// Piece size is set to 512 but we should edit how we do this on this receiver side to accept different piece sized based on peer

type Seeder struct {
	infoHash          []byte
	peerID            []byte
	connectedLeechers []Leecher
	filepath          string
}

type Leecher struct {
	infoHash []byte
	peerID   []byte
	bitfield []byte
	tcpConn  net.Conn
}

type SeederStack []Seeder

// In main, we should have a thread listening for new connections, that also has a SeederStack keeping track of all of the files that we are seeding
// For every file we fully download, we should create a new Seeder that continually listens for new connections
func (s *SeederStack) listen(port string) {
	// Listen for new connections
	conn, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Println("Error listening:", err)
		return
	}
	for {
		tcpConn, err := conn.Accept()
		if err != nil {
			continue
		}

		leecher := Leecher{
			nil,
			nil,
			nil,
			tcpConn,
		}

		// Handle the handshake
		go s.handleConn(leecher)
	}
}

type HandshakeMessage struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

// First, they exchange a handshake exchanging info_hash and peer_id
func (s *SeederStack) handleConn(leecher Leecher) {
	// Receive initial handshake
	buf := make([]byte, 68)
	leecher.tcpConn.Read(buf)
	handshake := HandshakeMessage{
		string(buf[:19]),
		[20]byte{},
		[20]byte{},
	}
	copy(handshake.InfoHash[:], buf[28:48])
	copy(handshake.PeerID[:], buf[48:68])

	leecher.infoHash = handshake.InfoHash[:]
	leecher.peerID = handshake.PeerID[:]
	leecher.bitfield = make([]byte, 0)

	// Find a seeder with the same info_hash
	cseeder := Seeder{
		nil,
		nil,
		nil,
		"",
	}
	for _, seeder := range *s {
		if bytes.Equal(seeder.infoHash, handshake.InfoHash[:]) {
			// Add leecher to seeder's list of connected leechers
			seeder.connectedLeechers = append(seeder.connectedLeechers, leecher)
			cseeder = seeder
			break
		}
	}

	if cseeder.infoHash == nil {
		// No seeder found
		log.Println("No seeder found for info_hash", handshake.InfoHash)
		return
	}

	// Send handshake response
	handshakeResponse := HandshakeMessage{
		"BitTorrent protocol",
		handshake.InfoHash,
		handshake.PeerID,
	}
	leecher.tcpConn.Write([]byte(handshakeResponse.Pstr))

	// Now we handle the rest of the messages
	for {
		buf := make([]byte, 4)
		leecher.tcpConn.Read(buf)
		length := int(buf[0])<<24 | int(buf[1])<<16 | int(buf[2])<<8 | int(buf[3])

		buf = make([]byte, length)
		leecher.tcpConn.Read(buf)

		// Here, we could stretch implement a switch statement to handle different types of messages that are in the actual protocol
		// For now, we handle just one, the bitfield message
		if buf[0] == 5 {
			// Handle bitfield message
			leecher.bitfield = buf[1:]
			cseeder.handleBitfield(leecher)
		}
	}
}

// Then, the leecher sends a bitfield message indicating which pieces it has
func (s *Seeder) handleBitfield(leecher Leecher) {
	// Send all of the pieces that the leecher doesn't have
	for _, b := range leecher.bitfield {
		for i := 0; i < 8; i++ {
			if b&(1<<i) == 0 {
				// Send piece
				s.sendPiece(i, leecher)
			}
		}
	}

}

// The seeder then responds by providing the pieces requested
func (s *Seeder) sendPiece(pieceIndex int, leecher Leecher) {
	// Open the file (we should keep open file and file descriptor in seeder struct)
	os.Open(s.filepath)
	// Seek to the correct position
	offset := int64(pieceIndex * PIECE_SIZE)
	file, err := os.Open(s.filepath)
	if err != nil {
		log.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	_, err = file.Seek(offset, 0)
	if err != nil {
		log.Println("Error seeking file:", err)
		return
	}
	// Read the piece
	buf := make([]byte, PIECE_SIZE)
	_, err = file.Read(buf)
	if err != nil {
		log.Println("Error reading file:", err)
		return
	}

	buf = append([]byte{byte(pieceIndex)}, buf...)
	// Send the piece
	leecher.tcpConn.Write(buf)
}
