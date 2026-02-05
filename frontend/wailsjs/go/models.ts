export namespace main {

	export class DisqusSetting {
		api: string;
		apikey: string;
		shortname: string;

		static createFrom(source: any = {}) {
			return new DisqusSetting(source);
		}

		constructor(source: any = {}) {
			if ('string' === typeof source) source = JSON.parse(source);
			this.api = source["api"];
			this.apikey = source["apikey"];
			this.shortname = source["shortname"];
		}
	}
	export class GitalkSetting {
		clientId: string;
		clientSecret: string;
		repository: string;
		owner: string;

		static createFrom(source: any = {}) {
			return new GitalkSetting(source);
		}

		constructor(source: any = {}) {
			if ('string' === typeof source) source = JSON.parse(source);
			this.clientId = source["clientId"];
			this.clientSecret = source["clientSecret"];
			this.repository = source["repository"];
			this.owner = source["owner"];
		}
	}



}



export class CommentSetting {
	enable: boolean;
	platform: string;
	platformConfigs: Record<string, any>;

	static createFrom(source: any = {}) {
		return new CommentSetting(source);
	}

	constructor(source: any = {}) {
		if ('string' === typeof source) source = JSON.parse(source);
		this.enable = source["enable"];
		this.platform = source["platform"];
		this.platformConfigs = source["platformConfigs"];
	}
}

export class Setting {
	platform: string;
	domain: string;
	repository: string;
	branch: string;
	username: string;
	email: string;
	tokenUsername: string;
	token: string;
	cname: string;
	port: string;
	server: string;
	password: string;
	privateKey: string;
	remotePath: string;
	proxyPath: string;
	proxyPort: string;
	enabledProxy: string;
	netlifySiteId: string;
	netlifyAccessToken: string;

	static createFrom(source: any = {}) {
		return new Setting(source);
	}

	constructor(source: any = {}) {
		if ('string' === typeof source) source = JSON.parse(source);
		this.platform = source["platform"];
		this.domain = source["domain"];
		this.repository = source["repository"];
		this.branch = source["branch"];
		this.username = source["username"];
		this.email = source["email"];
		this.tokenUsername = source["tokenUsername"];
		this.token = source["token"];
		this.cname = source["cname"];
		this.port = source["port"];
		this.server = source["server"];
		this.password = source["password"];
		this.privateKey = source["privateKey"];
		this.remotePath = source["remotePath"];
		this.proxyPath = source["proxyPath"];
		this.proxyPort = source["proxyPort"];
		this.enabledProxy = source["enabledProxy"];
		this.netlifySiteId = source["netlifySiteId"];
		this.netlifyAccessToken = source["netlifyAccessToken"];
	}
}
export class ThemeConfig {
	themeName: string;
	postPageSize: number;
	archivesPageSize: number;
	siteName: string;
	siteDescription: string;
	footerInfo: string;
	showFeatureImage: boolean;
	domain: string;
	postUrlFormat: string;
	tagUrlFormat: string;
	dateFormat: string;
	feedFullText: boolean;
	feedCount: number;
	archivesPath: string;
	postPath: string;
	tagPath: string;

	static createFrom(source: any = {}) {
		return new ThemeConfig(source);
	}

	constructor(source: any = {}) {
		if ('string' === typeof source) source = JSON.parse(source);
		this.themeName = source["themeName"];
		this.postPageSize = source["postPageSize"];
		this.archivesPageSize = source["archivesPageSize"];
		this.siteName = source["siteName"];
		this.siteDescription = source["siteDescription"];
		this.footerInfo = source["footerInfo"];
		this.showFeatureImage = source["showFeatureImage"];
		this.domain = source["domain"];
		this.postUrlFormat = source["postUrlFormat"];
		this.tagUrlFormat = source["tagUrlFormat"];
		this.dateFormat = source["dateFormat"];
		this.feedFullText = source["feedFullText"];
		this.feedCount = source["feedCount"];
		this.archivesPath = source["archivesPath"];
		this.postPath = source["postPath"];
		this.tagPath = source["tagPath"];
	}
}
export class SiteData {
	appDir: string;
	posts: any[];
	tags: any[];
	menus: any[];
	categories: any[];
	themeConfig: ThemeConfig;
	themeCustomConfig: Record<string, any>;
	currentThemeConfig: any[];
	themes: any[];
	setting: Setting;
	commentSetting: CommentSetting;

	static createFrom(source: any = {}) {
		return new SiteData(source);
	}

	constructor(source: any = {}) {
		if ('string' === typeof source) source = JSON.parse(source);
		this.appDir = source["appDir"];
		this.posts = source["posts"];
		this.tags = source["tags"];
		this.menus = source["menus"];
		this.categories = source["categories"];
		this.themeConfig = this.convertValues(source["themeConfig"], ThemeConfig);
		this.themeCustomConfig = source["themeCustomConfig"];
		this.currentThemeConfig = source["currentThemeConfig"];
		this.themes = source["themes"];
		this.setting = this.convertValues(source["setting"], Setting);
		this.commentSetting = this.convertValues(source["commentSetting"], CommentSetting);
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



