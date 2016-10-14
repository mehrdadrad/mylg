/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
import * as cdAst from '../expression_parser/ast';
import * as o from '../output/output_ast';
export interface NameResolver {
    callPipe(name: string, input: o.Expression, args: o.Expression[]): o.Expression;
    getLocal(name: string): o.Expression;
    createLiteralArray(values: o.Expression[]): o.Expression;
    createLiteralMap(values: Array<Array<string | o.Expression>>): o.Expression;
}
export declare class ExpressionWithWrappedValueInfo {
    expression: o.Expression;
    needsValueUnwrapper: boolean;
    temporaryCount: number;
    constructor(expression: o.Expression, needsValueUnwrapper: boolean, temporaryCount: number);
}
export declare function convertCdExpressionToIr(nameResolver: NameResolver, implicitReceiver: o.Expression, expression: cdAst.AST, valueUnwrapper: o.ReadVarExpr, bindingIndex: number): ExpressionWithWrappedValueInfo;
export declare function convertCdStatementToIr(nameResolver: NameResolver, implicitReceiver: o.Expression, stmt: cdAst.AST, bindingIndex: number): o.Statement[];
export declare function temporaryDeclaration(bindingIndex: number, temporaryNumber: number): o.Statement;
