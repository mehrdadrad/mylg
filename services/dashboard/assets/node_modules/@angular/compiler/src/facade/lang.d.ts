
/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
export interface BrowserNodeGlobal {
    Object: typeof Object;
    Array: typeof Array;
    Map: typeof Map;
    Set: typeof Set;
    Date: DateConstructor;
    RegExp: RegExpConstructor;
    JSON: typeof JSON;
    Math: any;
    assert(condition: any): void;
    Reflect: any;
    getAngularTestability: Function;
    getAllAngularTestabilities: Function;
    getAllAngularRootElements: Function;
    frameworkStabilizers: Array<Function>;
    setTimeout: Function;
    clearTimeout: Function;
    setInterval: Function;
    clearInterval: Function;
    encodeURI: Function;
}
export declare function scheduleMicroTask(fn: Function): void;
declare var _global: BrowserNodeGlobal;
export { _global as global };
export declare function getTypeNameForDebugging(type: any): string;
export declare function isPresent(obj: any): boolean;
export declare function isBlank(obj: any): boolean;
export declare function isBoolean(obj: any): boolean;
export declare function isNumber(obj: any): boolean;
export declare function isString(obj: any): obj is string;
export declare function isFunction(obj: any): boolean;
export declare function isType(obj: any): boolean;
export declare function isStringMap(obj: any): obj is Object;
export declare function isStrictStringMap(obj: any): boolean;
export declare function isArray(obj: any): boolean;
export declare function isDate(obj: any): obj is Date;
export declare function noop(): void;
export declare function stringify(token: any): string;
export declare class StringWrapper {
    static fromCharCode(code: number): string;
    static charCodeAt(s: string, index: number): number;
    static split(s: string, regExp: RegExp): string[];
    static equals(s: string, s2: string): boolean;
    static stripLeft(s: string, charVal: string): string;
    static stripRight(s: string, charVal: string): string;
    static replace(s: string, from: string, replace: string): string;
    static replaceAll(s: string, from: RegExp, replace: string): string;
    static slice<T>(s: string, from?: number, to?: number): string;
    static replaceAllMapped(s: string, from: RegExp, cb: (m: string[]) => string): string;
    static contains(s: string, substr: string): boolean;
    static compare(a: string, b: string): number;
}
export declare class StringJoiner {
    parts: string[];
    constructor(parts?: string[]);
    add(part: string): void;
    toString(): string;
}
export declare class NumberWrapper {
    static toFixed(n: number, fractionDigits: number): string;
    static equal(a: number, b: number): boolean;
    static parseIntAutoRadix(text: string): number;
    static parseInt(text: string, radix: number): number;
    static NaN: number;
    static isNumeric(value: any): boolean;
    static isNaN(value: any): boolean;
    static isInteger(value: any): boolean;
}
export declare var RegExp: RegExpConstructor;
export declare class FunctionWrapper {
    static apply(fn: Function, posArgs: any): any;
    static bind(fn: Function, scope: any): Function;
}
export declare function looseIdentical(a: any, b: any): boolean;
export declare function getMapKey<T>(value: T): T;
export declare function normalizeBlank(obj: Object): any;
export declare function normalizeBool(obj: boolean): boolean;
export declare function isJsObject(o: any): boolean;
export declare function print(obj: Error | Object): void;
export declare function warn(obj: Error | Object): void;
export declare class Json {
    static parse(s: string): Object;
    static stringify(data: Object): string;
}
export declare function setValueOnPath(global: any, path: string, value: any): void;
export declare function getSymbolIterator(): string | symbol;
export declare function evalExpression(sourceUrl: string, expr: string, declarations: string, vars: {
    [key: string]: any;
}): any;
export declare function isPrimitive(obj: any): boolean;
export declare function hasConstructor(value: Object, type: any): boolean;
export declare function escape(s: string): string;
export declare function escapeRegExp(s: string): string;
