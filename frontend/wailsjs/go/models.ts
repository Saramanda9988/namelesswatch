export namespace appconf {
	
	export class AppConfig {
	    ai_provider: string;
	    ai_base_url: string;
	    ai_model: string;
	    ai_token?: string;
	    ai_context_recent_turns: number;
	    ai_context_compact_turns: number;
	    ai_context_soft_budget: number;
	    ai_context_hard_budget: number;
	    ai_choice_prefetch_enabled: boolean;
	    ai_choice_prefetch_global_concurrency: number;
	    ai_choice_prefetch_session_concurrency: number;
	    ai_choice_prefetch_ttl_ms: number;
	    ai_choice_prefetch_wait_ms: number;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ai_provider = source["ai_provider"];
	        this.ai_base_url = source["ai_base_url"];
	        this.ai_model = source["ai_model"];
	        this.ai_token = source["ai_token"];
	        this.ai_context_recent_turns = source["ai_context_recent_turns"];
	        this.ai_context_compact_turns = source["ai_context_compact_turns"];
	        this.ai_context_soft_budget = source["ai_context_soft_budget"];
	        this.ai_context_hard_budget = source["ai_context_hard_budget"];
	        this.ai_choice_prefetch_enabled = source["ai_choice_prefetch_enabled"];
	        this.ai_choice_prefetch_global_concurrency = source["ai_choice_prefetch_global_concurrency"];
	        this.ai_choice_prefetch_session_concurrency = source["ai_choice_prefetch_session_concurrency"];
	        this.ai_choice_prefetch_ttl_ms = source["ai_choice_prefetch_ttl_ms"];
	        this.ai_choice_prefetch_wait_ms = source["ai_choice_prefetch_wait_ms"];
	    }
	}

}

export namespace roleplay {
	
	export class AchievementRule {
	    kind: string;
	    endingId?: string;
	    endingKind?: string;
	    forbidSnapshotFork?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AchievementRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.endingId = source["endingId"];
	        this.endingKind = source["endingKind"];
	        this.forbidSnapshotFork = source["forbidSnapshotFork"];
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
	export class AchievementDefinition {
	    id: string;
	    title: string;
	    type?: string;
	    trigger?: string;
	    requiresCustomInput?: boolean;
	    ending: Ending;
	    rule?: AchievementRule;
	
	    static createFrom(source: any = {}) {
	        return new AchievementDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.type = source["type"];
	        this.trigger = source["trigger"];
	        this.requiresCustomInput = source["requiresCustomInput"];
	        this.ending = this.convertValues(source["ending"], Ending);
	        this.rule = this.convertValues(source["rule"], AchievementRule);
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
	export class AchievementReference {
	    id: string;
	    title: string;
	
	    static createFrom(source: any = {}) {
	        return new AchievementReference(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	    }
	}
	
	export class AchievementUnlock {
	    gameId: string;
	    achievementId: string;
	    title: string;
	    sessionId: string;
	    endingId?: string;
	    unlockedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new AchievementUnlock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.gameId = source["gameId"];
	        this.achievementId = source["achievementId"];
	        this.title = source["title"];
	        this.sessionId = source["sessionId"];
	        this.endingId = source["endingId"];
	        this.unlockedAt = source["unlockedAt"];
	    }
	}
	export class AchievementUnlockResult {
	    gameId: string;
	    achievementId: string;
	    title: string;
	    sessionId: string;
	    endingId?: string;
	    unlockedAt?: string;
	    new: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AchievementUnlockResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.gameId = source["gameId"];
	        this.achievementId = source["achievementId"];
	        this.title = source["title"];
	        this.sessionId = source["sessionId"];
	        this.endingId = source["endingId"];
	        this.unlockedAt = source["unlockedAt"];
	        this.new = source["new"];
	    }
	}
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
	    customInput?: boolean;
	    tools?: ChoiceTool[];
	    scene?: SceneChange;
	    bgm?: BGMChange;
	    ending?: Ending;
	    achievement?: AchievementReference;
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
	        this.customInput = source["customInput"];
	        this.tools = this.convertValues(source["tools"], ChoiceTool);
	        this.scene = this.convertValues(source["scene"], SceneChange);
	        this.bgm = this.convertValues(source["bgm"], BGMChange);
	        this.ending = this.convertValues(source["ending"], Ending);
	        this.achievement = this.convertValues(source["achievement"], AchievementReference);
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
	    achievement?: AchievementUnlockResult;
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
	        this.achievement = this.convertValues(source["achievement"], AchievementUnlockResult);
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
	    achievements?: AchievementDefinition[];
	
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
	        this.achievements = this.convertValues(source["achievements"], AchievementDefinition);
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

