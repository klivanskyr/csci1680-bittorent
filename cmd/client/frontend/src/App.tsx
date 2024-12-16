import { Tab } from "./types"
import { Sidebar } from "./components"

import "./App.css"
import { Download, Home, Upload } from "./Pages"

export default function App() {
    const tabs: Tab[] = ["Home", "Download", "Upload"]

    return (
        <div id="App">
            <Sidebar tabs={tabs}>
                {(currentTab) => (
                    <>
                        {currentTab === "Home" && <Home />}
                        {currentTab === "Download" && <Download />}
                        {currentTab === "Upload" && <Upload />}
                    </>
                )}
            </Sidebar>
        </div>
    )
}