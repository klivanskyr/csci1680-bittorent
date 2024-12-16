import { useState } from "react"
import { Tab } from "./types"
import { FileSelect } from "./components"

import "./App.css"

export default function App() {
    const [tab, setTab] = useState<Tab>("download")

    return (
        <div id="App">
            <div className="row">
                <button className={tab === "download" ? "button-1 button-1-active" : "button-1"} onClick={() => setTab("download")}>Download</button>
                <button className={tab === "upload" ? "button-1 button-1-active" : "button-1"} onClick={() => setTab("upload")}>Upload</button>
            </div>
            <div className="col">
                <h1>{tab === "download" ? "Upload Torrent File" : "Upload a File"}</h1>
                <FileSelect tab={tab} />
            </div>
        </div>
    )
}