import "./FileSelect.css";
import { Tab } from "../../types";
import { ReadFileToBytes, SelectTorrentFile, HashInfo, UnmarshalTorrent, SelectAnyFile, SendTrackerRequest } from "../../../wailsjs/go/main/App";

export default function FileSelect({ tab }: { tab: Tab }) {

    const handleFileSelect = async () => {
        if (tab === "download") {
            // Parse torrent file
            const file = await SelectTorrentFile();
            const bytes = await ReadFileToBytes(file.Path);
            const torrent = await UnmarshalTorrent(bytes);

            const hash = await HashInfo(file.Path); // I know this is ugly but ill refactor later hopefully
            console.log("hash:", hash);

            // Start GET requests to tracker server
            const response = await SendTrackerRequest(torrent, hash);
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