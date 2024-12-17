import "./FileSelect.css";
import { Tab } from "../../types";
import { 
    ReadFileToBytes, 
    SelectTorrentFile, 
    UnmarshalTorrent,
    SelectAnyFile,
    SendTrackerRequest, 
    DownloadFromSeeders, 
    GeneratePeerID,
    CreateTorrentFile,
    SaveFileFromBytes,
} from "../../../wailsjs/go/main/App";
import { useState } from "react";

type File = {
    bytes: number[];
    name: string;
}

export default function FileSelect({ tab }: { tab: Tab }) {
    const [uploadedFile, setUploadedFile] = useState<File | null>(null); // used for uploading
    const [downloadedFile, setDownloadedFile] = useState<File | null>(null); // used for downloading

    const handleFileSelect = async () => {
        if (tab === "Download") {
            // Parse torrent file
            const file = await SelectTorrentFile();
            const bytes = await ReadFileToBytes(file.Path);
            const torrent = await UnmarshalTorrent(bytes);
            console.log("torrent:", torrent);

            const totalPieces = Math.ceil((torrent.Info.Length + torrent.Info.PieceLength - 1) / torrent.Info.PieceLength);
            console.log("totalPieces:", totalPieces);

            const peerId = await GeneratePeerID(); // I dont like how this is frontend

            // Start GET requests to tracker server
            const peers = await SendTrackerRequest(torrent, peerId);
            console.log("peers:", peers);

            // Start downloading file from peers
            const downloadedBytes = await DownloadFromSeeders(peers, torrent, totalPieces);
            setDownloadedFile({ bytes: downloadedBytes, name: torrent.Info.Name });


        } else if (tab === "Upload" ) { // tab === "upload"
            const file = await SelectAnyFile();
            const torrentBytes = await CreateTorrentFile(file.Path);
            setUploadedFile({ bytes: torrentBytes, name: file.Name });
        }
    }

    const handleDownload = async (tab: Tab) => {
        if (tab === "Upload") {
            // Save Torrent File
            if (!uploadedFile) return;
            await SaveFileFromBytes(uploadedFile!.bytes, uploadedFile!.name, "Torrent Files", "*.torrent");
        } else if (tab === "Download") {
            // Save File
            if (!downloadedFile) return
            await SaveFileFromBytes(downloadedFile!.bytes, downloadedFile!.name, "Downloaded Files", "*.*");
        }
    };

    return (
        <div>
            {((tab === "Download" && !downloadedFile) || (tab === "Upload" && !uploadedFile)) && <button className="button-1" onClick={() => handleFileSelect()}>Select File</button>}
            {(tab === "Upload" && uploadedFile) &&
                <button className="button-1 button-download" onClick={() => handleDownload("Upload")}>Download Torrent File</button>
            }
            {(tab === "Download" && downloadedFile) && 
                <button className="button-1 button-download" onClick={() => handleDownload("Download")}>Download File</button>
            }
        </div>
    )
}