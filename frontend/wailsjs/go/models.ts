export namespace main {
	
	export class MonthSummary {
	    year: number;
	    month: number;
	    earnings: number;
	    expenses: number;
	    earnings_by_category: Record<string, number>;
	    expenses_by_category: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new MonthSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.year = source["year"];
	        this.month = source["month"];
	        this.earnings = source["earnings"];
	        this.expenses = source["expenses"];
	        this.earnings_by_category = source["earnings_by_category"];
	        this.expenses_by_category = source["expenses_by_category"];
	    }
	}
	export class MonthTotal {
	    earnings: number;
	    expenses: number;
	
	    static createFrom(source: any = {}) {
	        return new MonthTotal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.earnings = source["earnings"];
	        this.expenses = source["expenses"];
	    }
	}
	export class Transaction {
	    id: string;
	    date: string;
	    type: string;
	    category: string;
	    description: string;
	    amount: number;
	
	    static createFrom(source: any = {}) {
	        return new Transaction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.date = source["date"];
	        this.type = source["type"];
	        this.category = source["category"];
	        this.description = source["description"];
	        this.amount = source["amount"];
	    }
	}
	export class MonthTransactions {
	    year: number;
	    month: number;
	    transactions: Transaction[];
	
	    static createFrom(source: any = {}) {
	        return new MonthTransactions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.year = source["year"];
	        this.month = source["month"];
	        this.transactions = this.convertValues(source["transactions"], Transaction);
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
	export class YearTotal {
	    earnings: number;
	    expenses: number;
	
	    static createFrom(source: any = {}) {
	        return new YearTotal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.earnings = source["earnings"];
	        this.expenses = source["expenses"];
	    }
	}
	export class OverallSummary {
	    earnings: number;
	    expenses: number;
	    years: Record<string, YearTotal>;
	
	    static createFrom(source: any = {}) {
	        return new OverallSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.earnings = source["earnings"];
	        this.expenses = source["expenses"];
	        this.years = this.convertValues(source["years"], YearTotal, true);
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
	
	export class UpdateInfo {
	    available: boolean;
	    version: string;
	    body: string;
	    downloadUrl: string;
	    assetName: string;
	    assetId: number;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.version = source["version"];
	        this.body = source["body"];
	        this.downloadUrl = source["downloadUrl"];
	        this.assetName = source["assetName"];
	        this.assetId = source["assetId"];
	    }
	}
	export class YearSummary {
	    year: number;
	    earnings: number;
	    expenses: number;
	    months: Record<string, MonthTotal>;
	    earnings_by_category: Record<string, number>;
	    expenses_by_category: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new YearSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.year = source["year"];
	        this.earnings = source["earnings"];
	        this.expenses = source["expenses"];
	        this.months = this.convertValues(source["months"], MonthTotal, true);
	        this.earnings_by_category = source["earnings_by_category"];
	        this.expenses_by_category = source["expenses_by_category"];
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

