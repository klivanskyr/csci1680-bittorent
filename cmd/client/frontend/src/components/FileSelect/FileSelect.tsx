import "./FileSelect.css";
import { Tab } from "../../types";
import { ReadFileToBytes, SelectTorrentFile, HashInfo, UnmarshalTorrent, SelectAnyFile, SendTrackerRequest, DownloadFromSeeders, GeneratePeerID } from "../../../wailsjs/go/main/App";

export default function FileSelect({ tab }: { tab: Tab }) {

    const handleFileSelect = async () => {
        if (tab === "download") {
            // Parse torrent file
            const file = await SelectTorrentFile();
            const bytes = await ReadFileToBytes(file.Path);
            const torrent = await UnmarshalTorrent(bytes);
            console.log("torrent:", torrent);

            // Find total pieces THIS IS ONLY SUPPORTED FOR SINGLE FILE TORRENTS
            const totalPieces = Math.ceil((torrent?.info?.length + torrent?.info?.["piece length"] - 1) / torrent?.info?.["piece length"]);

            console.log("totalPieces:", totalPieces);

            const infoHash = await HashInfo(file.Path); // I know this is ugly but ill refactor later hopefully
            const peerId = await GeneratePeerID(); // I dont like how this is all frontend

            // Start GET requests to tracker server
            const peers = await SendTrackerRequest(torrent, infoHash, peerId);
            console.log("peers:", peers);

            // Start downloading file from peers
            await DownloadFromSeeders(peers, infoHash, peerId, totalPieces);


        } else {
            const file = await SelectAnyFile();
            console.log(file);
        }
    }

    return (
        <div>
            <button className="button-1" onClick={() => handleFileSelect()}>Select File</button>
        </div>
    )
}