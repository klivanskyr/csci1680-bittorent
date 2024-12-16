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
    const [file, setFile] = useState<File | null>(null); // used for uploading

    const handleFileSelect = async () => {
        if (tab === "download") {
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
            await DownloadFromSeeders(peers, torrent, totalPieces);


        } else { // tab === "upload"
            const file = await SelectAnyFile();
            const torrentBytes = await CreateTorrentFile(file.Path);
            setFile({ bytes: torrentBytes, name: file.Name });
        }
    }

    const handleDownload = async () => {
        if (!file) return;
        await SaveFileFromBytes(file.bytes, file.name, "Torrent Files", "*.torrent"); 
    };

    return (
        <div>
            {(tab === "download" || !file) && <button className="button-1" onClick={() => handleFileSelect()}>Select File</button>}
            {(tab === "upload" && file) &&
                <button className="button-1 button-download" onClick={() => handleDownload()}>Download Torrent File</button>
            }
        </div>
    )
}