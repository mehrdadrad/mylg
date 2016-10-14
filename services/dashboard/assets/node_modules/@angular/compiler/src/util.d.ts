/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
import { CompileTokenMetadata } from './compile_metadata';
import * as o from './output/output_ast';
export declare const MODULE_SUFFIX: string;
export declare function camelCaseToDashCase(input: string): string;
export declare function splitAtColon(input: string, defaultValues: string[]): string[];
export declare function splitAtPeriod(input: string, defaultValues: string[]): string[];
export declare function sanitizeIdentifier(name: string): string;
export declare function visitValue(value: any, visitor: ValueVisitor, context: any): any;
export interface ValueVisitor {
    visitArray(arr: any[], context: any): any;
    visitStringMap(map: {
        [key: string]: any;
    }, context: any): any;
    visitPrimitive(value: any, context: any): any;
    visitOther(value: any, context: any): any;
}
export declare class ValueTransformer implements ValueVisitor {
    visitArray(arr: any[], context: any): any;
    visitStringMap(map: {
        [key: string]: any;
    }, context: any): any;
    visitPrimitive(value: any, context: any): any;
    visitOther(value: any, context: any): any;
}
export declare function assetUrl(pkg: string, path?: string, type?: string): string;
export declare function createDiTokenExpression(token: CompileTokenMetadata): o.Expression;
export declare class SyncAsyncResult<T> {
    syncResult: T;
    asyncResult: Promise<T>;
    constructor(syncResult: T, asyncResult?: Promise<T>);
}
