export namespace backend {
	
	export class FileInfo {
	    Path: string;
	    Name: string;
	    Size: number;
	    Length: number;
	    // Go type: time
	    ModTime: any;
	    IsDir: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Name = source["Name"];
	        this.Size = source["Size"];
	        this.Length = source["Length"];
	        this.ModTime = this.convertValues(source["ModTime"], null);
	        this.IsDir = source["IsDir"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace torrent {
	
	export class TorrentInfo {
	    Name: string;
	    Length: number;
	    PieceLength: number;
	    Pieces: number[];
	
	    static createFrom(source: any = {}) {
	        return new TorrentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Length = source["Length"];
	        this.PieceLength = source["PieceLength"];
	        this.Pieces = source["Pieces"];
	    }
	}
	export class Torrent {
	    Announce: string;
	    Info: TorrentInfo;
	
	    static createFrom(source: any = {}) {
	        return new Torrent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Announce = source["Announce"];
	        this.Info = this.convertValues(source["Info"], TorrentInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace trackingserver {
	
	export class Peer {
	    PeerID: string;
	    Seeder: boolean;
	    IP: string;
	    Port: number;
	    // Go type: time
	    LastAnnounce: any;
	
	    static createFrom(source: any = {}) {
	        return new Peer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PeerID = source["PeerID"];
	        this.Seeder = source["Seeder"];
	        this.IP = source["IP"];
	        this.Port = source["Port"];
	        this.LastAnnounce = this.convertValues(source["LastAnnounce"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

