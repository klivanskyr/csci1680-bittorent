import "./FileSelect.css";
import { Tab } from "../../types";
import { SelectTorrentFile, SelectAnyFile, ReadFileToBytes, UnmarshalTorrent, SendTrackerRequest } from "../../../wailsjs/go/main/App";

export default function FileSelect({ tab }: { tab: Tab }) {

    const handleFileSelect = async () => {
        if (tab === "download") {
            // Parse torrent file
            const file = await SelectTorrentFile();
            const bytes = await ReadFileToBytes(file.Path);
            const torrent = await UnmarshalTorrent(bytes);
            console.log(torrent);

            // // Start GET requests to tracker server
            const response = await SendTrackerRequest(torrent);
            console.log("response:", response);

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