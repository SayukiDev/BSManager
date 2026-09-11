export namespace options {
	
	export class SecondInstanceData {
	    Args: string[];
	    WorkingDirectory: string;
	
	    static createFrom(source: any = {}) {
	        return new SecondInstanceData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Args = source["Args"];
	        this.WorkingDirectory = source["WorkingDirectory"];
	    }
	}

}

export namespace service {
	
	export class BaseStation {
	    name: string;
	    addr: string;
	    rssi: number;
	
	    static createFrom(source: any = {}) {
	        return new BaseStation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.addr = source["addr"];
	        this.rssi = source["rssi"];
	    }
	}
	export class Status {
	    setuped: boolean;
	    powerOn: boolean;
	    followSteamVR: boolean;
	    steamVRRunning: boolean;
	    managerRunning: boolean;
	    trackerCount: number;
	    devices: settings.BaseStation[];
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.setuped = source["setuped"];
	        this.powerOn = source["powerOn"];
	        this.followSteamVR = source["followSteamVR"];
	        this.steamVRRunning = source["steamVRRunning"];
	        this.managerRunning = source["managerRunning"];
	        this.trackerCount = source["trackerCount"];
	        this.devices = this.convertValues(source["devices"], settings.BaseStation);
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

export namespace settings {
	
	export class BaseStation {
	    enable: boolean;
	    name: string;
	    addr: string;
	
	    static createFrom(source: any = {}) {
	        return new BaseStation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enable = source["enable"];
	        this.name = source["name"];
	        this.addr = source["addr"];
	    }
	}
	export class Content {
	    setuped: boolean;
	    logLevel: string;
	    devices: BaseStation[];
	    checkInterval: number;
	    followSteamVR: boolean;
	    checkTrackerConnected: boolean;
	    minTrackerCount: number;
	    shutdownWaiting: number;
	
	    static createFrom(source: any = {}) {
	        return new Content(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.setuped = source["setuped"];
	        this.logLevel = source["logLevel"];
	        this.devices = this.convertValues(source["devices"], BaseStation);
	        this.checkInterval = source["checkInterval"];
	        this.followSteamVR = source["followSteamVR"];
	        this.checkTrackerConnected = source["checkTrackerConnected"];
	        this.minTrackerCount = source["minTrackerCount"];
	        this.shutdownWaiting = source["shutdownWaiting"];
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

