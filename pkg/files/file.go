package backend

import (
	"context"
	"os"
	"time"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

type FileInfo struct {
    Path    string
    Name    string
    Size    int64
    Length  int64
    ModTime time.Time // Last modification time
    IsDir   bool      // Whether it's a directory
}

// SelectTorrentFile opens a file dialog, allowing only .torrent files and returns the selected file path
func SelectTorrentFile(ctx context.Context) (*FileInfo, error) {
    filePath, err := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
        Title: "Select a Torrent File",
        Filters: []runtime.FileFilter{
            {
                DisplayName: "Torrent Files",
                Pattern:     "*.torrent",
            },
        },
    })

    if err != nil || filePath == "" {
        return nil, err
    }

    // Get file information
    info, err := os.Stat(filePath)
    if err != nil {
        return nil, err
    }

    return &FileInfo{
        Path:    filePath,
        Length:  info.Size(),
        Name:    info.Name(),
        Size:    info.Size(),
        ModTime: info.ModTime(),
        IsDir:   info.IsDir(),
    }, nil
}

// SelectAnyFile opens a file dialog, allowing any file type and returns the selected file path
func SelectAnyFile(ctx context.Context) (*FileInfo, error) {
    filePath, err := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
        Title: "Select a File",
    })

    if err != nil || filePath == "" {
        return nil, err
    }

    // Get file information
    info, err := os.Stat(filePath)
    if err != nil {
        return nil, err
    }

    return &FileInfo{
        Path:    filePath,
        Length:  info.Size(),
        Name:    info.Name(),
        Size:    info.Size(),
        ModTime: info.ModTime(),
        IsDir:   info.IsDir(),
    }, nil
}

func ReadFileToBytes(path string) ([]byte, error) {
    return os.ReadFile(path)
}

func SaveFileFromBytes(ctx context.Context, data []byte, defaultFileName string, displayName string, pattern string) error {
	// Open save file dialog
	savePath, err := runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{
		Title: defaultFileName,
		Filters: []runtime.FileFilter{
			{
				DisplayName: displayName,
				Pattern:     pattern,
			},
		},
	})

	if err != nil || savePath == "" {
		return err
	}

	// Write the data to the selected path
	return os.WriteFile(savePath, data, 0644)
}
