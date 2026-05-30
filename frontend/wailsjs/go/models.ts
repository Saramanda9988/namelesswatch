export namespace appconf {
	
	export class AppConfig {
	    ai_provider: string;
	    ai_base_url: string;
	    ai_model: string;
	    ai_token?: string;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ai_provider = source["ai_provider"];
	        this.ai_base_url = source["ai_base_url"];
	        this.ai_model = source["ai_model"];
	        this.ai_token = source["ai_token"];
	    }
	}

}

export namespace roleplay {
	
	export class BGMAsset {
	    id: string;
	    name?: string;
	    fileName: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new BGMAsset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.fileName = source["fileName"];
	        this.url = source["url"];
	    }
	}
	export class BGMChange {
	    action: string;
	    id?: string;
	    reason?: string;
	
	    static createFrom(source: any = {}) {
	        return new BGMChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.id = source["id"];
	        this.reason = source["reason"];
	    }
	}
	export class ChoiceOption {
	    id: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new ChoiceOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	    }
	}
	export class ChoiceTool {
	    type: string;
	    id: string;
	    prompt?: string;
	    options: ChoiceOption[];
	
	    static createFrom(source: any = {}) {
	        return new ChoiceTool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.id = source["id"];
	        this.prompt = source["prompt"];
	        this.options = this.convertValues(source["options"], ChoiceOption);
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
	export class Ending {
	    id: string;
	    title: string;
	    kind: string;
	
	    static createFrom(source: any = {}) {
	        return new Ending(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.kind = source["kind"];
	    }
	}
	export class SceneChange {
	    id: string;
	    reason?: string;
	
	    static createFrom(source: any = {}) {
	        return new SceneChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.reason = source["reason"];
	    }
	}
	export class GameTurn {
	    id: string;
	    role: string;
	    payload: string[];
	    selectedChoiceId?: string;
	    selectedChoiceLabel?: string;
	    tools?: ChoiceTool[];
	    scene?: SceneChange;
	    bgm?: BGMChange;
	    ending?: Ending;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new GameTurn(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.role = source["role"];
	        this.payload = source["payload"];
	        this.selectedChoiceId = source["selectedChoiceId"];
	        this.selectedChoiceLabel = source["selectedChoiceLabel"];
	        this.tools = this.convertValues(source["tools"], ChoiceTool);
	        this.scene = this.convertValues(source["scene"], SceneChange);
	        this.bgm = this.convertValues(source["bgm"], BGMChange);
	        this.ending = this.convertValues(source["ending"], Ending);
	        this.createdAt = source["createdAt"];
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
	export class GameSession {
	    id: string;
	    gameId: string;
	    state: string;
	    currentSceneId?: string;
	    currentBgmId?: string;
	    workspacePath: string;
	    memoryPath: string;
	    turns: GameTurn[];
	    label?: string;
	    isSnapshot?: boolean;
	    parentId?: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new GameSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.gameId = source["gameId"];
	        this.state = source["state"];
	        this.currentSceneId = source["currentSceneId"];
	        this.currentBgmId = source["currentBgmId"];
	        this.workspacePath = source["workspacePath"];
	        this.memoryPath = source["memoryPath"];
	        this.turns = this.convertValues(source["turns"], GameTurn);
	        this.label = source["label"];
	        this.isSnapshot = source["isSnapshot"];
	        this.parentId = source["parentId"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
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
	
	export class GameTurnResult {
	    sessionId: string;
	    gameId: string;
	    state: string;
	    payload: string[];
	    tools: ChoiceTool[];
	    scene?: SceneChange;
	    bgm?: BGMChange;
	    currentBgmId?: string;
	    ending?: Ending;
	    turn: GameTurn;
	
	    static createFrom(source: any = {}) {
	        return new GameTurnResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.gameId = source["gameId"];
	        this.state = source["state"];
	        this.payload = source["payload"];
	        this.tools = this.convertValues(source["tools"], ChoiceTool);
	        this.scene = this.convertValues(source["scene"], SceneChange);
	        this.bgm = this.convertValues(source["bgm"], BGMChange);
	        this.currentBgmId = source["currentBgmId"];
	        this.ending = this.convertValues(source["ending"], Ending);
	        this.turn = this.convertValues(source["turn"], GameTurn);
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
	export class SceneAsset {
	    id: string;
	    name: string;
	    fileName: string;
	    url: string;
	    x: number;
	    y: number;
	    hasPosition: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SceneAsset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.fileName = source["fileName"];
	        this.url = source["url"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.hasPosition = source["hasPosition"];
	    }
	}
	export class LibraryGame {
	    id: string;
	    title: string;
	    importedAt: string;
	    files: Record<string, string>;
	    photoUrls: string[];
	    mapUrls: string[];
	    scenes?: SceneAsset[];
	    bgms?: BGMAsset[];
	
	    static createFrom(source: any = {}) {
	        return new LibraryGame(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.importedAt = source["importedAt"];
	        this.files = source["files"];
	        this.photoUrls = source["photoUrls"];
	        this.mapUrls = source["mapUrls"];
	        this.scenes = this.convertValues(source["scenes"], SceneAsset);
	        this.bgms = this.convertValues(source["bgms"], BGMAsset);
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
	export class ImportGameResult {
	    game?: LibraryGame;
	    missing: string[];
	    warnings: string[];
	    validFiles: string[];
	
	    static createFrom(source: any = {}) {
	        return new ImportGameResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.game = this.convertValues(source["game"], LibraryGame);
	        this.missing = source["missing"];
	        this.warnings = source["warnings"];
	        this.validFiles = source["validFiles"];
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

export namespace service {
	
	export class SessionSummary {
	    id: string;
	    gameId: string;
	    state: string;
	    label?: string;
	    isSnapshot: boolean;
	    turnCount: number;
	    preview: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.gameId = source["gameId"];
	        this.state = source["state"];
	        this.label = source["label"];
	        this.isSnapshot = source["isSnapshot"];
	        this.turnCount = source["turnCount"];
	        this.preview = source["preview"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}

}

