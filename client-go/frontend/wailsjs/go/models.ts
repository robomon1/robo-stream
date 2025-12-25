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
	export class Button {
	    id: string;
	    row: number;
	    col: number;
	    text: string;
	    color: string;
	    icon?: string;
	    action: ButtonAction;
	
	    static createFrom(source: any = {}) {
	        return new Button(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.row = source["row"];
	        this.col = source["col"];
	        this.text = source["text"];
	        this.color = source["color"];
	        this.icon = source["icon"];
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
	export class ButtonConfig {
	    grid: GridConfig;
	    buttons: Button[];
	
	    static createFrom(source: any = {}) {
	        return new ButtonConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.grid = this.convertValues(source["grid"], GridConfig);
	        this.buttons = this.convertValues(source["buttons"], Button);
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

