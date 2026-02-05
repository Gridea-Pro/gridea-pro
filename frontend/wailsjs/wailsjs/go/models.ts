export namespace domain {
	
	export class Category {
	    name: string;
	    slug: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new Category(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.slug = source["slug"];
	        this.description = source["description"];
	    }
	}
	export class Comment {
	    id: string;
	    avatar: string;
	    nickname: string;
	    email?: string;
	    url?: string;
	    content: string;
	    createdAt: string;
	    articleId: string;
	    articleTitle: string;
	    articleUrl?: string;
	    parentId?: string;
	    parentNick?: string;
	    isNew?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Comment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.avatar = source["avatar"];
	        this.nickname = source["nickname"];
	        this.email = source["email"];
	        this.url = source["url"];
	        this.content = source["content"];
	        this.createdAt = source["createdAt"];
	        this.articleId = source["articleId"];
	        this.articleTitle = source["articleTitle"];
	        this.articleUrl = source["articleUrl"];
	        this.parentId = source["parentId"];
	        this.parentNick = source["parentNick"];
	        this.isNew = source["isNew"];
	    }
	}
	export class CommentSettings {
	    enable: boolean;
	    platform: string;
	    platformConfigs: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new CommentSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enable = source["enable"];
	        this.platform = source["platform"];
	        this.platformConfigs = source["platformConfigs"];
	    }
	}
	export class FileInfo {
	    name: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	    }
	}
	export class Link {
	    id: string;
	    name: string;
	    url: string;
	    description: string;
	    avatar: string;
	
	    static createFrom(source: any = {}) {
	        return new Link(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.description = source["description"];
	        this.avatar = source["avatar"];
	    }
	}
	export class Menu {
	    name: string;
	    link: string;
	    openType: string;
	
	    static createFrom(source: any = {}) {
	        return new Menu(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.link = source["link"];
	        this.openType = source["openType"];
	    }
	}
	export class PaginatedComments {
	    comments: Comment[];
	    total: number;
	    page: number;
	    pageSize: number;
	    totalPages: number;
	
	    static createFrom(source: any = {}) {
	        return new PaginatedComments(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.comments = this.convertValues(source["comments"], Comment);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.totalPages = source["totalPages"];
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
	export class PostData {
	    title: string;
	    date: string;
	    tags: string[];
	    tagIds: string[];
	    categories: string[];
	    published: boolean;
	    hideInList: boolean;
	    feature: string;
	    isTop: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PostData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.date = source["date"];
	        this.tags = source["tags"];
	        this.tagIds = source["tagIds"];
	        this.categories = source["categories"];
	        this.published = source["published"];
	        this.hideInList = source["hideInList"];
	        this.feature = source["feature"];
	        this.isTop = source["isTop"];
	    }
	}
	export class Post {
	    data: PostData;
	    content: string;
	    abstract: string;
	    fileName: string;
	
	    static createFrom(source: any = {}) {
	        return new Post(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], PostData);
	        this.content = source["content"];
	        this.abstract = source["abstract"];
	        this.fileName = source["fileName"];
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
	
	export class PostInput {
	    title: string;
	    date: string;
	    tags: string[];
	    tagIds: string[];
	    categories: string[];
	    published: boolean;
	    hideInList: boolean;
	    isTop: boolean;
	    content: string;
	    fileName: string;
	    deleteFileName: string;
	    featureImage: FileInfo;
	    featureImagePath: string;
	
	    static createFrom(source: any = {}) {
	        return new PostInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.date = source["date"];
	        this.tags = source["tags"];
	        this.tagIds = source["tagIds"];
	        this.categories = source["categories"];
	        this.published = source["published"];
	        this.hideInList = source["hideInList"];
	        this.isTop = source["isTop"];
	        this.content = source["content"];
	        this.fileName = source["fileName"];
	        this.deleteFileName = source["deleteFileName"];
	        this.featureImage = this.convertValues(source["featureImage"], FileInfo);
	        this.featureImagePath = source["featureImagePath"];
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
	export class Tag {
	    id: string;
	    name: string;
	    slug: string;
	    used: boolean;
	    color?: string;
	
	    static createFrom(source: any = {}) {
	        return new Tag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.slug = source["slug"];
	        this.used = source["used"];
	        this.color = source["color"];
	    }
	}
	export class Theme {
	    folder: string;
	    name: string;
	    version: string;
	    description?: string;
	    author?: string;
	    repository?: string;
	    previewImage?: string;
	    customConfig?: any[];
	
	    static createFrom(source: any = {}) {
	        return new Theme(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.folder = source["folder"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.description = source["description"];
	        this.author = source["author"];
	        this.repository = source["repository"];
	        this.previewImage = source["previewImage"];
	        this.customConfig = source["customConfig"];
	    }
	}
	export class ThemeConfig {
	    themeName: string;
	    postPageSize: number;
	    archivesPageSize: number;
	    siteName: string;
	    siteAuthor: string;
	    siteEmail: string;
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
	    linkPath: string;
	    customConfig?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new ThemeConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.themeName = source["themeName"];
	        this.postPageSize = source["postPageSize"];
	        this.archivesPageSize = source["archivesPageSize"];
	        this.siteName = source["siteName"];
	        this.siteAuthor = source["siteAuthor"];
	        this.siteEmail = source["siteEmail"];
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
	        this.linkPath = source["linkPath"];
	        this.customConfig = source["customConfig"];
	    }
	}
	export class UploadedFile {
	    name: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new UploadedFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	    }
	}

}

export namespace facade {
	
	export class RendererFacade {
	
	
	    static createFrom(source: any = {}) {
	        return new RendererFacade(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

