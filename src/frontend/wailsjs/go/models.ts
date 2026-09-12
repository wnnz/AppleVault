export namespace backend {
	
	export class AccountInfo {
	    id: string;
	    name: string;
	    email: string;
	    region: string;
	    active: boolean;
	    success: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AccountInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.region = source["region"];
	        this.active = source["active"];
	        this.success = source["success"];
	    }
	}
	export class AppItem {
	    id: number;
	    bundleID: string;
	    name: string;
	    version: string;
	    price: number;
	    purchaseDate?: string;
	    displayPrice: string;
	
	    static createFrom(source: any = {}) {
	        return new AppItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.bundleID = source["bundleID"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.price = source["price"];
	        this.purchaseDate = source["purchaseDate"];
	        this.displayPrice = source["displayPrice"];
	    }
	}
	export class DeviceInfo {
	    udid: string;
	    name: string;
	    productType: string;
	    productVersion: string;
	    connectionType: string;
	
	    static createFrom(source: any = {}) {
	        return new DeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.udid = source["udid"];
	        this.name = source["name"];
	        this.productType = source["productType"];
	        this.productVersion = source["productVersion"];
	        this.connectionType = source["connectionType"];
	    }
	}
	export class DownloadResult {
	    output: string;
	    purchased: boolean;
	    success: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DownloadResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.output = source["output"];
	        this.purchased = source["purchased"];
	        this.success = source["success"];
	    }
	}
	export class DownloadTask {
	    id: string;
	    accountID?: string;
	    appName: string;
	    bundleID: string;
	    appId: number;
	    version: string;
	    versionId: string;
	    fileSize: string;
	    totalBytes: number;
	    currBytes: number;
	    progress: number;
	    speed: string;
	    status: string;
	    outputPath: string;
	    errorMessage: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new DownloadTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.accountID = source["accountID"];
	        this.appName = source["appName"];
	        this.bundleID = source["bundleID"];
	        this.appId = source["appId"];
	        this.version = source["version"];
	        this.versionId = source["versionId"];
	        this.fileSize = source["fileSize"];
	        this.totalBytes = source["totalBytes"];
	        this.currBytes = source["currBytes"];
	        this.progress = source["progress"];
	        this.speed = source["speed"];
	        this.status = source["status"];
	        this.outputPath = source["outputPath"];
	        this.errorMessage = source["errorMessage"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class InstallResult {
	    success: boolean;
	    message: string;
	    output: string;
	
	    static createFrom(source: any = {}) {
	        return new InstallResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.output = source["output"];
	    }
	}
	export class IPAInspectionResult {
	    success: boolean;
	    appName: string;
	    bundleID: string;
	    version: string;
	    buildVersion: string;
	    minimumOSVersion: string;
	    supportedPlatforms: string[];
	    supportedDeviceTypes: string[];
	    signed: boolean;
	    fileSize: number;
	    displayFileSize: string;
	    compatible: boolean;
	    compatibilityMessage: string;

	    static createFrom(source: any = {}) {
	        return new IPAInspectionResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.appName = source["appName"];
	        this.bundleID = source["bundleID"];
	        this.version = source["version"];
	        this.buildVersion = source["buildVersion"];
	        this.minimumOSVersion = source["minimumOSVersion"];
	        this.supportedPlatforms = source["supportedPlatforms"];
	        this.supportedDeviceTypes = source["supportedDeviceTypes"];
	        this.signed = source["signed"];
	        this.fileSize = source["fileSize"];
	        this.displayFileSize = source["displayFileSize"];
	        this.compatible = source["compatible"];
	        this.compatibilityMessage = source["compatibilityMessage"];
	    }
	}
	export class LoginResult {
	    success: boolean;
	    requires2FA: boolean;
	    account: AccountInfo;
	    errorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new LoginResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.requires2FA = source["requires2FA"];
	        this.account = this.convertValues(source["account"], AccountInfo);
	        this.errorMessage = source["errorMessage"];
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
	export class ProxyTestResult {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class PurchaseResult {
	    alreadyOwned: boolean;
	    success: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PurchaseResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.alreadyOwned = source["alreadyOwned"];
	        this.success = source["success"];
	    }
	}
	export class PurchasedResult {
	    count: number;
	    totalCount: number;
	    page: number;
	    apps: AppItem[];
	
	    static createFrom(source: any = {}) {
	        return new PurchasedResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.totalCount = source["totalCount"];
	        this.page = source["page"];
	        this.apps = this.convertValues(source["apps"], AppItem);
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
	export class SearchResult {
	    count: number;
	    apps: AppItem[];
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.apps = this.convertValues(source["apps"], AppItem);
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
	export class Settings {
	    keychainPassphrase: string;
	    defaultDownloadDir: string;
	    defaultPlatform: string;
	    enableProxy: boolean;
	    proxyUrl: string;
	    ipaToolPath: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.keychainPassphrase = source["keychainPassphrase"];
	        this.defaultDownloadDir = source["defaultDownloadDir"];
	        this.defaultPlatform = source["defaultPlatform"];
	        this.enableProxy = source["enableProxy"];
	        this.proxyUrl = source["proxyUrl"];
	        this.ipaToolPath = source["ipaToolPath"];
	    }
	}
	export class VersionMetadataResult {
	    externalVersionID: string;
	    displayVersion: string;
	    releaseDate: string;
	    fileSize: number;
	    displayFileSize: string;
	    success: boolean;
	
	    static createFrom(source: any = {}) {
	        return new VersionMetadataResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.externalVersionID = source["externalVersionID"];
	        this.displayVersion = source["displayVersion"];
	        this.releaseDate = source["releaseDate"];
	        this.fileSize = source["fileSize"];
	        this.displayFileSize = source["displayFileSize"];
	        this.success = source["success"];
	    }
	}
	export class UpdateInfo {
	    currentVersion: string;
	    latestVersion: string;
	    available: boolean;
	    releaseName: string;
	    releaseNotes: string;
	    releaseURL: string;
	    publishedAt: string;

	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.available = source["available"];
	        this.releaseName = source["releaseName"];
	        this.releaseNotes = source["releaseNotes"];
	        this.releaseURL = source["releaseURL"];
	        this.publishedAt = source["publishedAt"];
	    }
	}
	export class VersionsResult {
	    bundleID: string;
	    externalVersionIdentifiers: string[];
	    success: boolean;
	
	    static createFrom(source: any = {}) {
	        return new VersionsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bundleID = source["bundleID"];
	        this.externalVersionIdentifiers = source["externalVersionIdentifiers"];
	        this.success = source["success"];
	    }
	}

}

