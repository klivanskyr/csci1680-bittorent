import "./Upload.css";
import { FileSelect } from "../../components";

export default function Upload() {
    return (
        <div className="col">
            <h1>Upload any file to get started.</h1>
            <FileSelect tab={"Upload"} />
        </div>
    )
}