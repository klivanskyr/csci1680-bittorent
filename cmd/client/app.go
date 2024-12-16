package main

import (
	"context"

	"bittorrent/pkg/client"
	"bittorrent/pkg/files"
	Torrent "bittorrent/pkg/torrent"
	"bittorrent/pkg/trackingserver"
)

// App struct
type App struct {
	ctx context.Context
	seederStack *Torrent.SeederStack
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Create a new App	
	a := App{}

	// start seeder stack
	a.seederStack = &Torrent.SeederStack{}
	go a.seederStack.Listen(6881, 10) // Start listening on port 6881 with 10 retries

	return &a
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectTorrentFile opens a file dialog, allowing only .torrent files and returns the selected file path
func (a *App) SelectTorrentFile() (*backend.FileInfo, error) {
    return backend.SelectTorrentFile(a.ctx)
}

// SelectAnyFile opens a file dialog, allowing any file type and returns the selected file path
func (a *App) SelectAnyFile() (*backend.FileInfo, error) {
    return backend.SelectAnyFile(a.ctx)
}

// ReadFile reads a file and returns the file data
func (a *App) ReadFileToBytes(path string) ([]byte, error) {
	return backend.ReadFileToBytes(path)
}

// ConvertBencodeToJSON converts bencoded data to JSON
func (a *App) UnmarshalTorrent(data []byte) (*Torrent.Torrent, error) {
	t, err := Torrent.UnmarshalTorrent(data)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// SendTrackerRequest sends a GET request to the tracker's announce URL
func (a *App) SendTrackerRequest(torrent Torrent.Torrent, peerId string) ([]trackingserver.Peer, error) {
	return client.SendTrackerRequest(torrent, peerId)
}

func (a *App) HashInfo(torrent Torrent.Torrent) ([]byte, error) {
	return torrent.HashInfo()
}

func (a *App) DownloadFromSeeders(peers []trackingserver.Peer, torrent Torrent.Torrent, totalPieces uint32) error {
	return Torrent.DownloadFromSeeders(peers, torrent, totalPieces)
}

func (a *App) GeneratePeerID() string {
	return client.GeneratePeerID()
}

func (a *App) CreateTorrentFile(filePath string) ([]byte, error) {
	return Torrent.CreateTorrentFile(a.seederStack, filePath, client.GeneratePeerID()) // Every torrent file has new peerId which is wrong
}

func (a *App) SaveFileFromBytes(data []byte, defaultFileName string, displayName string, pattern string) error {
	return backend.SaveFileFromBytes(a.ctx, data, defaultFileName, displayName, pattern)
}
