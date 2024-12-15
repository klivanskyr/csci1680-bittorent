import "./FileSelect.css";
import { Tab } from "../../types";
import { SelectTorrentFile, SelectAnyFile, ReadFileToBytes, UnmarshalTorrent, SendTrackerRequest, GeneratePeerID } from "../../../wailsjs/go/main/App";

export default function FileSelect({ tab }: { tab: Tab }) {

    const handleFileSelect = async () => {
        if (tab === "download") {
            // Parse torrent file
            const file = await SelectTorrentFile();
            const bytes = await ReadFileToBytes(file.Path);
            const torrent = await UnmarshalTorrent(bytes);
            console.log(torrent);

            // Start GET requests to tracker server
            const peerID = await GeneratePeerID();
            const response = await SendTrackerRequest(torrent, peerID);
            

        } else {
            const file = await SelectAnyFile();
        }
    }

    return (
        <div>
            <button className="button-1" onClick={() => handleFileSelect()}>Select File</button>
        </div>
    )
}