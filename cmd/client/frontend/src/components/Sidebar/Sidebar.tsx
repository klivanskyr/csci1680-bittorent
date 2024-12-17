import "./Sidebar.css";
import { useState } from "react"
import { Tab } from "../../types";

type SidebarProps = {
    tabs: Tab[];
    children: (currentTab: Tab) => React.ReactNode;
};

export default function Sidebar({ tabs, children }: SidebarProps) {
    const [currentTab, setCurrentTab] = useState<Tab>("Home");

    return (
        <div className="row">
            <div className="sidebar">
                {tabs.map((tab, i) => (
                    <div key={i} className={`item ${currentTab === tab ? "active-item" : ""}`} onClick={() => setCurrentTab(tab)}>
                        <button  className="item-button">{String(tab)}</button>
                    </div>
                ))}
            </div>
            <div className="content">
                {children(currentTab)}
            </div>
        </div>
    )
}