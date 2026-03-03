export namespace controller {
	
	export class ParamDef {
	    key: string;
	    label: string;
	    type: string;
	    options?: string[];
	    required?: boolean;
	    default?: any;
	
	    static createFrom(source: any = {}) {
	        return new ParamDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.options = source["options"];
	        this.required = source["required"];
	        this.default = source["default"];
	    }
	}
	export class ActionTypeDefinition {
	    type: string;
	    name: string;
	    description?: string;
	    params?: ParamDef[];
	    indicator_field?: string;
	    indicator_invert?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ActionTypeDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.params = this.convertValues(source["params"], ParamDef);
	        this.indicator_field = source["indicator_field"];
	        this.indicator_invert = source["indicator_invert"];
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
	export class ConfigField {
	    key: string;
	    label: string;
	    type: string;
	    required?: boolean;
	    default?: string;
	    help?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigField(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.required = source["required"];
	        this.default = source["default"];
	        this.help = source["help"];
	    }
	}

}

export namespace main {
	
	export class ControllerActionGroup {
	    controller_id: string;
	    controller_name: string;
	    connected: boolean;
	    actions: controller.ActionTypeDefinition[];
	
	    static createFrom(source: any = {}) {
	        return new ControllerActionGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.controller_id = source["controller_id"];
	        this.controller_name = source["controller_name"];
	        this.connected = source["connected"];
	        this.actions = this.convertValues(source["actions"], controller.ActionTypeDefinition);
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

export namespace models {
	
	export class ButtonAction {
	    controller?: string;
	    type: string;
	    params?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new ButtonAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.controller = source["controller"];
	        this.type = source["type"];
	        this.params = source["params"];
	    }
	}
	export class Button {
	    id: string;
	    name: string;
	    description: string;
	    icon: string;
	    color: string;
	    action: ButtonAction;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Button(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.action = this.convertValues(source["action"], ButtonAction);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	
	export class ClientSession {
	    session_id: string;
	    client_id: string;
	    client_name: string;
	    config_id: string;
	    ip_address: string;
	    // Go type: time
	    last_connected: any;
	    // Go type: time
	    last_active: any;
	
	    static createFrom(source: any = {}) {
	        return new ClientSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.session_id = source["session_id"];
	        this.client_id = source["client_id"];
	        this.client_name = source["client_name"];
	        this.config_id = source["config_id"];
	        this.ip_address = source["ip_address"];
	        this.last_connected = this.convertValues(source["last_connected"], null);
	        this.last_active = this.convertValues(source["last_active"], null);
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
	export class GridConfig {
	    rows: number;
	    cols: number;
	
	    static createFrom(source: any = {}) {
	        return new GridConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rows = source["rows"];
	        this.cols = source["cols"];
	    }
	}
	export class Configuration {
	    id: string;
	    name: string;
	    description: string;
	    grid: GridConfig;
	    buttons: Record<string, string>;
	    is_default: boolean;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Configuration(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.grid = this.convertValues(source["grid"], GridConfig);
	        this.buttons = source["buttons"];
	        this.is_default = source["is_default"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	
	export class ResolvedButton {
	    id: string;
	    row: number;
	    col: number;
	    text: string;
	    icon: string;
	    color: string;
	    action: ButtonAction;
	
	    static createFrom(source: any = {}) {
	        return new ResolvedButton(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.row = source["row"];
	        this.col = source["col"];
	        this.text = source["text"];
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.action = this.convertValues(source["action"], ButtonAction);
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
	export class ResolvedConfiguration {
	    id: string;
	    name: string;
	    grid: GridConfig;
	    buttons: ResolvedButton[];
	
	    static createFrom(source: any = {}) {
	        return new ResolvedConfiguration(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.grid = this.convertValues(source["grid"], GridConfig);
	        this.buttons = this.convertValues(source["buttons"], ResolvedButton);
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

