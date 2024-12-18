package torrent

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"

	"bittorrent/pkg/trackingserver"

	"github.com/zeebo/bencode"
)

// All the seeder needs to do is respond to requests for specific pieces of a file

// Piece size is set to 512 but we should edit how we do this on this receiver side to accept different piece sized based on peer

type Seeder struct {
	addr              net.Addr
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

type SeederStack struct {
	mtx     sync.Mutex
	seeders []Seeder
	port    int
}

// Important Constants
// const TrackerAddr string = "http://20.121.67.21:80/announce"
// const TrackerAddr string = "http://localhost:8080/announce"
const TrackerAddr string = "http://[fd95:13b9:7f7:7322:c26c:53d:5bc7:5487]:8080/announce" // My localhost IPv6 address

// Adds Seeder to SeederStack and sends POST request to tracker
func (s *SeederStack) AddSeeder(seeder Seeder) error {
	s.mtx.Lock()
	s.seeders = append(s.seeders, seeder)
	s.mtx.Unlock()

	// Get ipv6 address
	host, _, err := net.SplitHostPort(seeder.addr.String())
	if err != nil {
		return fmt.Errorf("failed to parse seeder address: %v", err)
	}

	// Send a request to the tracker to announce the seeder
	announce := trackingserver.AnnounceRequest{
		InfoHash: seeder.infoHash,
		PeerID:   seeder.peerID,
		IP:       host,
		Port:     s.port,
		Event:    trackingserver.STARTED,
	}

	// //debubg announce
	// fmt.Println("InfoHash: ", announce.InfoHash)
	// fmt.Println("PeerID: ", announce.PeerID)
	// fmt.Println("IP: ", announce.IP)
	// fmt.Println("Port: ", announce.Port)
	// fmt.Println("Event: ", announce.Event)

	var bencodedAnnounce bytes.Buffer
	err = bencode.NewEncoder(&bencodedAnnounce).Encode(announce)
	if err != nil {
		return err
	}

	response, err := http.Post(TrackerAddr, "application/x-bittorrent", &bencodedAnnounce)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Read the response body to bytes
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// Decode the response
	var decodedResponse map[string]interface{}
	err = bencode.NewDecoder(bytes.NewReader(responseBytes)).Decode(&decodedResponse)
	if err != nil {
		return err
	}

	fmt.Println("Decoded response:", decodedResponse)

	return nil
}

// In main, we should have a thread listening for new connections, that also has a SeederStack keeping track of all of the files that we are seeding
// For every file we fully download, we should create a new Seeder that continually listens for new connections
// Listen tries to bind to a port (string) and retries with consecutive ports up to a limit
func (s *SeederStack) Listen(startPort int, maxRetries int) {
	var listener net.Listener
	var err error
	currentPort := startPort

	// Retry listening on consecutive ports until maxRetries is reached
	for i := 0; i < maxRetries; i++ {
		portStr := strconv.Itoa(currentPort)
		listener, err = net.Listen("tcp", "[::]:"+portStr) // Supports IPv6
		if err == nil {
			log.Printf("Listening on port %d", currentPort)
			break
		}

		log.Printf("Port %d is unavailable, retrying with port %d...", currentPort, currentPort+1)
		currentPort++
	}

	// If no ports were available, log an error and return
	if err != nil {
		log.Fatalf("Error: Unable to bind to any port after %d retries: %v", maxRetries, err)
		return
	}

	s.mtx.Lock()
	s.port = currentPort
	s.mtx.Unlock()

	defer listener.Close()

	fmt.Println("Listening on port", currentPort)

	// Accept incoming connections
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		leecher := Leecher{
			nil,
			nil,
			nil,
			tcpConn,
		}

		// Handle the connection
		fmt.Println("Received connection from", tcpConn.RemoteAddr())
		go s.handleConn(leecher)
	}
}

type HandshakeMessage struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func (hm *HandshakeMessage) Marshal() ([]byte, error) {
	// Marshal the handshake message
	buf := make([]byte, 68)
	buf[0] = byte(len(hm.Pstr))
	copy(buf[1:], []byte(hm.Pstr))

	copy(buf[28:], hm.InfoHash[:])
	copy(buf[48:], hm.PeerID[:])

	return buf, nil
}

func UnmarshalHandshake(buf []byte) (*HandshakeMessage, error) {
	// Unmarshal the handshake message
	hm := HandshakeMessage{
		string(buf[1:20]),
		[20]byte{},
		[20]byte{},
	}
	copy(hm.InfoHash[:], buf[28:48])
	copy(hm.PeerID[:], buf[48:68])

	return &hm, nil
}

// First, they exchange a handshake exchanging info_hash and peer_id
func (s *SeederStack) handleConn(leecher Leecher) {
	fmt.Println("Handling connection from", leecher.tcpConn.RemoteAddr())
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
		nil,
		"",
	}

	s.mtx.Lock()
	for _, seeder := range s.seeders {
		if bytes.Equal(seeder.infoHash, handshake.InfoHash[:]) {
			// Add leecher to seeder's list of connected leechers
			seeder.connectedLeechers = append(seeder.connectedLeechers, leecher)
			cseeder = seeder
			break
		}
	}
	s.mtx.Unlock()

	if cseeder.infoHash == nil {
		// No seeder found
		log.Println("No seeder found for info_hash", handshake.InfoHash)
		return
	}

	fmt.Println("Seeder found for info_hash", hex.EncodeToString(handshake.InfoHash[:]))



	// Send handshake response
	handshakeResponse := HandshakeMessage{
		"BitTorrent protocol",
		handshake.InfoHash,
		handshake.PeerID,
	}
	handshakeresponseBytes, err := handshakeResponse.Marshal()
	if err != nil {
		log.Println("Error marshalling handshake response:", err)
		return
	}
	fmt.Println("Sending handshake response")
	leecher.tcpConn.Write(handshakeresponseBytes)
	fmt.Println("Handshake response sent")

	// Now we handle the rest of the messages
	var test = 0
	for {
		buf := make([]byte, 5)
		leecher.tcpConn.Read(buf)

		length := uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
		buf2 := make([]byte, length)
		leecher.tcpConn.Read(buf2)

		buf = append(buf, buf2...)

		// Unmarshal the message
		message, err := UnmarshalMessage(buf)
		if err != nil {
			log.Println("Error unmarshalling message:", err)
			return
		}	

		if test < 5 {
			fmt.Println("RECIEVED MESSAGE FROM LEECHER: ", message)
			test++
		}

		// Here, we could stretch implement a switch statement to handle different types of messages that are in the actual protocol
		// For now, we handle just one, the bitfield message
		if message.ID == 5 {
			// We don't handle this case in the end

			// Handle bitfield message
			// leecher.bitfield = buf[1:]
			// cseeder.handleBitfield(leecher)
		} else if message.ID == 6 {
			// Handle request message
			// The index is the first 4 bytes of the payload
			pieceIndex := uint32(message.Payload[0])<<24 | uint32(message.Payload[1])<<16 | uint32(message.Payload[2])<<8 | uint32(message.Payload[3])

			// cseeder.handleRequest(leecher, buf)
			cseeder.sendPiece(pieceIndex, leecher)
		}
	}
}

const (
	Choke         int8 = 0
	Unchoke       int8 = 1
	Interested    int8 = 2
	NotInterested int8 = 3
	Bitfield      int8 = 5
	Request       int8 = 6
	Piece         int8 = 7
)

type Message struct {
	Length  uint32
	ID      int8
	Payload []byte
}

func (m *Message) Marshal() ([]byte, error) {
	// Marshal the message
	buf := make([]byte, 5+len(m.Payload))
	buf[0] = byte(m.Length >> 24)
	buf[1] = byte(m.Length >> 16)
	buf[2] = byte(m.Length >> 8)
	buf[3] = byte(m.Length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)

	return buf, nil
}

func UnmarshalMessage(buf []byte) (*Message, error) {
	// Unmarshal the message
	m := Message{
		uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3]),
		int8(buf[4]),
		buf[5:],
	}

	return &m, nil
}

// The seeder then responds by providing the pieces requested
func (s *Seeder) sendPiece(pieceIndex uint32, leecher Leecher) {
	// Open the file (we should keep open file and file descriptor in seeder struct)
	file, err := os.Open(s.filepath)
	// Seek to the correct position
	offset := int64(pieceIndex * PIECE_SIZE)
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
	n, err := file.Read(buf) // n is the actual number of bytes read
	if err != nil && err != io.EOF {
		log.Println("Error reading file:", err)
		return
	}
	buf = buf[:n] // Trim buf to the actual number of bytes read

	// We convert piece index to 4 bytes
	pieceIndexBytes := uint32ToBytes(pieceIndex)

	// Now set the message length correctly
	message := Message{
		Length:  uint32(n + 5), // 1 byte for the ID + 4 byte for the piece index + n bytes for the piece
		ID:      7,
		Payload: append(pieceIndexBytes, buf...),
	}

	fmt.Println("Sending bytes: ", string(buf))
	fmt.Println("Sending bytes: ", message)

	// Marshal the message
	msgBytes, err := message.Marshal()
	if err != nil {
		log.Println("Error marshalling piece message:", err)
		return
	}

	// Send the piece
	fmt.Println("Sending piece", pieceIndex)
	leecher.tcpConn.Write(msgBytes)
}
