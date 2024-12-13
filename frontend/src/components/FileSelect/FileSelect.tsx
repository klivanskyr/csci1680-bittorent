import "./FileSelect.css";
import { Tab } from "../../types";
import { SelectTorrentFile, SelectAnyFile, ReadFileToBytes, UnmarshalTorrent } from "../../../wailsjs/go/main/App";

export default function FileSelect({ tab }: { tab: Tab }) {

    const handleFileSelect = async () => {
        if (tab === "download") {
            const file = await SelectTorrentFile();
            const bytes = await ReadFileToBytes(file.Path);
            const torrent = await UnmarshalTorrent(bytes);
            console.log(torrent);

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