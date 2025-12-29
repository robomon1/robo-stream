export namespace config {
	
	export class ButtonAction {
	    type: string;
	    params: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new ButtonAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.params = source["params"];
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

