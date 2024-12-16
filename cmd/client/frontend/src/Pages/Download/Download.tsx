import "./Download.css";
import { FileSelect } from "../../components";

export default function Download() {
    return (
        <div className="col">
            <h1>Upload a Torrent file to get started.</h1>
            <FileSelect tab={"Download"} />
        </div>
    )
}