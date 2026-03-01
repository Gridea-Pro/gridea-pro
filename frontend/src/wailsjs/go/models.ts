export namespace domain {
	
	export class Category {
	    _id: string;
	    name: string;
	    slug: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new Category(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this._id = source["_id"];
	        this.name = source["name"];
	        this.slug = source["slug"];
	        this.description = source["description"];
	    }
	}
	export class Comment {
	    id: string;
	    avatar: string;
	    nickname: string;
	    email: string;
	    url: string;
	    content: string;
	    createdAt: string;
	    articleId: string;
	    articleTitle: string;
	    articleUrl: string;
	    parentId: string;
	    parentNick: string;
	    isNew: boolean;
	
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
	    _id: string;
	    name: string;
	    url: string;
	    description: string;
	    avatar: string;
	
	    static createFrom(source: any = {}) {
	        return new Link(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this._id = source["_id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.description = source["description"];
	        this.avatar = source["avatar"];
	    }
	}
	export class Memo {
	    _id: string;
	    content: string;
	    tags: string[];
	    images: string[];
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Memo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this._id = source["_id"];
	        this.content = source["content"];
	        this.tags = source["tags"];
	        this.images = source["images"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class TagStat {
	    name: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new TagStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.count = source["count"];
	    }
	}
	export class MemoStats {
	    total: number;
	    tags: TagStat[];
	    heatmap: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new MemoStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.tags = this.convertValues(source["tags"], TagStat);
	        this.heatmap = source["heatmap"];
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
	export class MemoDashboardDTO {
	    memos: Memo[];
	    stats: MemoStats;
	
	    static createFrom(source: any = {}) {
	        return new MemoDashboardDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.memos = this.convertValues(source["memos"], Memo);
	        this.stats = this.convertValues(source["stats"], MemoStats);
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
	
	export class Menu {
	    _id: string;
	    name: string;
	    link: string;
	    openType: string;
	
	    static createFrom(source: any = {}) {
	        return new Menu(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this._id = source["_id"];
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
	export class Post {
	    id: string;
	    title: string;
	    createdAt: string;
	    updatedAt: string;
	    tags: string[];
	    tagIds: string[];
	    categories: string[];
	    categoryIds: string[];
	    published: boolean;
	    hideInList: boolean;
	    isTop: boolean;
	    feature: string;
	    featureImagePath: string;
	    featureImage: FileInfo;
	    content: string;
	    fileName: string;
	    deleteFileName: string;
	    abstract: string;
	
	    static createFrom(source: any = {}) {
	        return new Post(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.tags = source["tags"];
	        this.tagIds = source["tagIds"];
	        this.categories = source["categories"];
	        this.categoryIds = source["categoryIds"];
	        this.published = source["published"];
	        this.hideInList = source["hideInList"];
	        this.isTop = source["isTop"];
	        this.feature = source["feature"];
	        this.featureImagePath = source["featureImagePath"];
	        this.featureImage = this.convertValues(source["featureImage"], FileInfo);
	        this.content = source["content"];
	        this.fileName = source["fileName"];
	        this.deleteFileName = source["deleteFileName"];
	        this.abstract = source["abstract"];
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
	    _id: string;
	    name: string;
	    slug: string;
	    used: boolean;
	    color?: string;
	
	    static createFrom(source: any = {}) {
	        return new Tag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this._id = source["_id"];
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
	    domain: string;
	    postUrlFormat: string;
	    tagUrlFormat: string;
	    dateFormat: string;
	    language: string;
	    feedFullText: boolean;
	    feedCount: number;
	    postPath: string;
	    tagPath: string;
	    tagsPath: string;
	    linkPath: string;
	    memosPath: string;
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
	        this.domain = source["domain"];
	        this.postUrlFormat = source["postUrlFormat"];
	        this.tagUrlFormat = source["tagUrlFormat"];
	        this.dateFormat = source["dateFormat"];
	        this.language = source["language"];
	        this.feedFullText = source["feedFullText"];
	        this.feedCount = source["feedCount"];
	        this.postPath = source["postPath"];
	        this.tagPath = source["tagPath"];
	        this.tagsPath = source["tagsPath"];
	        this.linkPath = source["linkPath"];
	        this.memosPath = source["memosPath"];
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
	
	export class CategoryForm {
	    id: string;
	    name: string;
	    slug: string;
	    description: string;
	    originalSlug: string;
	
	    static createFrom(source: any = {}) {
	        return new CategoryForm(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.slug = source["slug"];
	        this.description = source["description"];
	        this.originalSlug = source["originalSlug"];
	    }
	}
	export class LinkForm {
	    id: string;
	    name: string;
	    url: string;
	    avatar: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new LinkForm(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.avatar = source["avatar"];
	        this.description = source["description"];
	    }
	}
	export class MenuForm {
	    name: string;
	    openType: string;
	    link: string;
	    index: any;
	
	    static createFrom(source: any = {}) {
	        return new MenuForm(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.openType = source["openType"];
	        this.link = source["link"];
	        this.index = source["index"];
	    }
	}
	export class PostDashboardDTO {
	    posts: domain.Post[];
	    tags: domain.Tag[];
	
	    static createFrom(source: any = {}) {
	        return new PostDashboardDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.posts = this.convertValues(source["posts"], domain.Post);
	        this.tags = this.convertValues(source["tags"], domain.Tag);
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
	export class PostForm {
	    id: string;
	    title: string;
	    createdAt: string;
	    tags: string[];
	    tagIds: string[];
	    categories: string[];
	    categoryIds: string[];
	    published: boolean;
	    hideInList: boolean;
	    isTop: boolean;
	    content: string;
	    fileName: string;
	    deleteFileName: string;
	    featureImage: domain.FileInfo;
	    featureImagePath: string;
	
	    static createFrom(source: any = {}) {
	        return new PostForm(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.createdAt = source["createdAt"];
	        this.tags = source["tags"];
	        this.tagIds = source["tagIds"];
	        this.categories = source["categories"];
	        this.categoryIds = source["categoryIds"];
	        this.published = source["published"];
	        this.hideInList = source["hideInList"];
	        this.isTop = source["isTop"];
	        this.content = source["content"];
	        this.fileName = source["fileName"];
	        this.deleteFileName = source["deleteFileName"];
	        this.featureImage = this.convertValues(source["featureImage"], domain.FileInfo);
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
	export class RendererFacade {
	
	
	    static createFrom(source: any = {}) {
	        return new RendererFacade(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class TagForm {
	    name: string;
	    slug: string;
	    color: string;
	    originalName: string;
	
	    static createFrom(source: any = {}) {
	        return new TagForm(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.slug = source["slug"];
	        this.color = source["color"];
	        this.originalName = source["originalName"];
	    }
	}

}

