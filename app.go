package main

import (
	"context"

	"bittorrent/backend"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
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
func (a *App) UnmarshalTorrent(data []byte) (interface{}, error) {
	return backend.UnmarshalTorrent(data)
}

// SendTrackerRequest sends a GET request to the tracker's announce URL
func (a *App) SendTrackerRequest(torrent interface{}, infoHash []byte, peerId string) ([]string, error) {
	return backend.SendTrackerRequest(torrent, infoHash, peerId)
}

func (a *App) HashInfo(torrentPath string) ([]byte, error) {
	return backend.HashInfo(torrentPath)
}

func (a *App) DownloadFromSeeders(peers []string, infoHash []byte, peerId string, totalPieces uint32) error {
	return backend.DownloadFromSeeders(peers, infoHash, peerId, totalPieces)
}

func (a *App) GeneratePeerID() string {
	return backend.GeneratePeerID()
}
