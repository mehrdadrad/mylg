/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
import * as cdAst from '../expression_parser/ast';
import { isArray, isBlank, isPresent } from '../facade/lang';
import { Identifiers, resolveIdentifier } from '../identifiers';
import * as o from '../output/output_ast';
export var ExpressionWithWrappedValueInfo = (function () {
    function ExpressionWithWrappedValueInfo(expression, needsValueUnwrapper, temporaryCount) {
        this.expression = expression;
        this.needsValueUnwrapper = needsValueUnwrapper;
        this.temporaryCount = temporaryCount;
    }
    return ExpressionWithWrappedValueInfo;
}());
export function convertCdExpressionToIr(nameResolver, implicitReceiver, expression, valueUnwrapper, bindingIndex) {
    var visitor = new _AstToIrVisitor(nameResolver, implicitReceiver, valueUnwrapper, bindingIndex);
    var irAst = expression.visit(visitor, _Mode.Expression);
    return new ExpressionWithWrappedValueInfo(irAst, visitor.needsValueUnwrapper, visitor.temporaryCount);
}
export function convertCdStatementToIr(nameResolver, implicitReceiver, stmt, bindingIndex) {
    var visitor = new _AstToIrVisitor(nameResolver, implicitReceiver, null, bindingIndex);
    var statements = [];
    flattenStatements(stmt.visit(visitor, _Mode.Statement), statements);
    prependTemporaryDecls(visitor.temporaryCount, bindingIndex, statements);
    return statements;
}
function temporaryName(bindingIndex, temporaryNumber) {
    return "tmp_" + bindingIndex + "_" + temporaryNumber;
}
export function temporaryDeclaration(bindingIndex, temporaryNumber) {
    return new o.DeclareVarStmt(temporaryName(bindingIndex, temporaryNumber), o.NULL_EXPR);
}
function prependTemporaryDecls(temporaryCount, bindingIndex, statements) {
    for (var i = temporaryCount - 1; i >= 0; i--) {
        statements.unshift(temporaryDeclaration(bindingIndex, i));
    }
}
var _Mode;
(function (_Mode) {
    _Mode[_Mode["Statement"] = 0] = "Statement";
    _Mode[_Mode["Expression"] = 1] = "Expression";
})(_Mode || (_Mode = {}));
function ensureStatementMode(mode, ast) {
    if (mode !== _Mode.Statement) {
        throw new Error("Expected a statement, but saw " + ast);
    }
}
function ensureExpressionMode(mode, ast) {
    if (mode !== _Mode.Expression) {
        throw new Error("Expected an expression, but saw " + ast);
    }
}
function convertToStatementIfNeeded(mode, expr) {
    if (mode === _Mode.Statement) {
        return expr.toStmt();
    }
    else {
        return expr;
    }
}
var _AstToIrVisitor = (function () {
    function _AstToIrVisitor(_nameResolver, _implicitReceiver, _valueUnwrapper, bindingIndex) {
        this._nameResolver = _nameResolver;
        this._implicitReceiver = _implicitReceiver;
        this._valueUnwrapper = _valueUnwrapper;
        this.bindingIndex = bindingIndex;
        this._nodeMap = new Map();
        this._resultMap = new Map();
        this._currentTemporary = 0;
        this.needsValueUnwrapper = false;
        this.temporaryCount = 0;
    }
    _AstToIrVisitor.prototype.visitBinary = function (ast, mode) {
        var op;
        switch (ast.operation) {
            case '+':
                op = o.BinaryOperator.Plus;
                break;
            case '-':
                op = o.BinaryOperator.Minus;
                break;
            case '*':
                op = o.BinaryOperator.Multiply;
                break;
            case '/':
                op = o.BinaryOperator.Divide;
                break;
            case '%':
                op = o.BinaryOperator.Modulo;
                break;
            case '&&':
                op = o.BinaryOperator.And;
                break;
            case '||':
                op = o.BinaryOperator.Or;
                break;
            case '==':
                op = o.BinaryOperator.Equals;
                break;
            case '!=':
                op = o.BinaryOperator.NotEquals;
                break;
            case '===':
                op = o.BinaryOperator.Identical;
                break;
            case '!==':
                op = o.BinaryOperator.NotIdentical;
                break;
            case '<':
                op = o.BinaryOperator.Lower;
                break;
            case '>':
                op = o.BinaryOperator.Bigger;
                break;
            case '<=':
                op = o.BinaryOperator.LowerEquals;
                break;
            case '>=':
                op = o.BinaryOperator.BiggerEquals;
                break;
            default:
                throw new Error("Unsupported operation " + ast.operation);
        }
        return convertToStatementIfNeeded(mode, new o.BinaryOperatorExpr(op, this.visit(ast.left, _Mode.Expression), this.visit(ast.right, _Mode.Expression)));
    };
    _AstToIrVisitor.prototype.visitChain = function (ast, mode) {
        ensureStatementMode(mode, ast);
        return this.visitAll(ast.expressions, mode);
    };
    _AstToIrVisitor.prototype.visitConditional = function (ast, mode) {
        var value = this.visit(ast.condition, _Mode.Expression);
        return convertToStatementIfNeeded(mode, value.conditional(this.visit(ast.trueExp, _Mode.Expression), this.visit(ast.falseExp, _Mode.Expression)));
    };
    _AstToIrVisitor.prototype.visitPipe = function (ast, mode) {
        var input = this.visit(ast.exp, _Mode.Expression);
        var args = this.visitAll(ast.args, _Mode.Expression);
        var value = this._nameResolver.callPipe(ast.name, input, args);
        this.needsValueUnwrapper = true;
        return convertToStatementIfNeeded(mode, this._valueUnwrapper.callMethod('unwrap', [value]));
    };
    _AstToIrVisitor.prototype.visitFunctionCall = function (ast, mode) {
        return convertToStatementIfNeeded(mode, this.visit(ast.target, _Mode.Expression).callFn(this.visitAll(ast.args, _Mode.Expression)));
    };
    _AstToIrVisitor.prototype.visitImplicitReceiver = function (ast, mode) {
        ensureExpressionMode(mode, ast);
        return this._implicitReceiver;
    };
    _AstToIrVisitor.prototype.visitInterpolation = function (ast, mode) {
        ensureExpressionMode(mode, ast);
        var args = [o.literal(ast.expressions.length)];
        for (var i = 0; i < ast.strings.length - 1; i++) {
            args.push(o.literal(ast.strings[i]));
            args.push(this.visit(ast.expressions[i], _Mode.Expression));
        }
        args.push(o.literal(ast.strings[ast.strings.length - 1]));
        return o.importExpr(resolveIdentifier(Identifiers.interpolate)).callFn(args);
    };
    _AstToIrVisitor.prototype.visitKeyedRead = function (ast, mode) {
        return convertToStatementIfNeeded(mode, this.visit(ast.obj, _Mode.Expression).key(this.visit(ast.key, _Mode.Expression)));
    };
    _AstToIrVisitor.prototype.visitKeyedWrite = function (ast, mode) {
        var obj = this.visit(ast.obj, _Mode.Expression);
        var key = this.visit(ast.key, _Mode.Expression);
        var value = this.visit(ast.value, _Mode.Expression);
        return convertToStatementIfNeeded(mode, obj.key(key).set(value));
    };
    _AstToIrVisitor.prototype.visitLiteralArray = function (ast, mode) {
        return convertToStatementIfNeeded(mode, this._nameResolver.createLiteralArray(this.visitAll(ast.expressions, mode)));
    };
    _AstToIrVisitor.prototype.visitLiteralMap = function (ast, mode) {
        var parts = [];
        for (var i = 0; i < ast.keys.length; i++) {
            parts.push([ast.keys[i], this.visit(ast.values[i], _Mode.Expression)]);
        }
        return convertToStatementIfNeeded(mode, this._nameResolver.createLiteralMap(parts));
    };
    _AstToIrVisitor.prototype.visitLiteralPrimitive = function (ast, mode) {
        return convertToStatementIfNeeded(mode, o.literal(ast.value));
    };
    _AstToIrVisitor.prototype.visitMethodCall = function (ast, mode) {
        var leftMostSafe = this.leftMostSafeNode(ast);
        if (leftMostSafe) {
            return this.convertSafeAccess(ast, leftMostSafe, mode);
        }
        else {
            var args = this.visitAll(ast.args, _Mode.Expression);
            var result = null;
            var receiver = this.visit(ast.receiver, _Mode.Expression);
            if (receiver === this._implicitReceiver) {
                var varExpr = this._nameResolver.getLocal(ast.name);
                if (isPresent(varExpr)) {
                    result = varExpr.callFn(args);
                }
            }
            if (isBlank(result)) {
                result = receiver.callMethod(ast.name, args);
            }
            return convertToStatementIfNeeded(mode, result);
        }
    };
    _AstToIrVisitor.prototype.visitPrefixNot = function (ast, mode) {
        return convertToStatementIfNeeded(mode, o.not(this.visit(ast.expression, _Mode.Expression)));
    };
    _AstToIrVisitor.prototype.visitPropertyRead = function (ast, mode) {
        var leftMostSafe = this.leftMostSafeNode(ast);
        if (leftMostSafe) {
            return this.convertSafeAccess(ast, leftMostSafe, mode);
        }
        else {
            var result = null;
            var receiver = this.visit(ast.receiver, _Mode.Expression);
            if (receiver === this._implicitReceiver) {
                result = this._nameResolver.getLocal(ast.name);
            }
            if (isBlank(result)) {
                result = receiver.prop(ast.name);
            }
            return convertToStatementIfNeeded(mode, result);
        }
    };
    _AstToIrVisitor.prototype.visitPropertyWrite = function (ast, mode) {
        var receiver = this.visit(ast.receiver, _Mode.Expression);
        if (receiver === this._implicitReceiver) {
            var varExpr = this._nameResolver.getLocal(ast.name);
            if (isPresent(varExpr)) {
                throw new Error('Cannot assign to a reference or variable!');
            }
        }
        return convertToStatementIfNeeded(mode, receiver.prop(ast.name).set(this.visit(ast.value, _Mode.Expression)));
    };
    _AstToIrVisitor.prototype.visitSafePropertyRead = function (ast, mode) {
        return this.convertSafeAccess(ast, this.leftMostSafeNode(ast), mode);
    };
    _AstToIrVisitor.prototype.visitSafeMethodCall = function (ast, mode) {
        return this.convertSafeAccess(ast, this.leftMostSafeNode(ast), mode);
    };
    _AstToIrVisitor.prototype.visitAll = function (asts, mode) {
        var _this = this;
        return asts.map(function (ast) { return _this.visit(ast, mode); });
    };
    _AstToIrVisitor.prototype.visitQuote = function (ast, mode) {
        throw new Error('Quotes are not supported for evaluation!');
    };
    _AstToIrVisitor.prototype.visit = function (ast, mode) {
        var result = this._resultMap.get(ast);
        if (result)
            return result;
        return (this._nodeMap.get(ast) || ast).visit(this, mode);
    };
    _AstToIrVisitor.prototype.convertSafeAccess = function (ast, leftMostSafe, mode) {
        // If the expression contains a safe access node on the left it needs to be converted to
        // an expression that guards the access to the member by checking the receiver for blank. As
        // execution proceeds from left to right, the left most part of the expression must be guarded
        // first but, because member access is left associative, the right side of the expression is at
        // the top of the AST. The desired result requires lifting a copy of the the left part of the
        // expression up to test it for blank before generating the unguarded version.
        // Consider, for example the following expression: a?.b.c?.d.e
        // This results in the ast:
        //         .
        //        / \
        //       ?.   e
        //      /  \
        //     .    d
        //    / \
        //   ?.  c
        //  /  \
        // a    b
        // The following tree should be generated:
        //
        //        /---- ? ----\
        //       /      |      \
        //     a   /--- ? ---\  null
        //        /     |     \
        //       .      .     null
        //      / \    / \
        //     .  c   .   e
        //    / \    / \
        //   a   b  ,   d
        //         / \
        //        .   c
        //       / \
        //      a   b
        //
        // Notice that the first guard condition is the left hand of the left most safe access node
        // which comes in as leftMostSafe to this routine.
        var guardedExpression = this.visit(leftMostSafe.receiver, _Mode.Expression);
        var temporary;
        if (this.needsTemporary(leftMostSafe.receiver)) {
            // If the expression has method calls or pipes then we need to save the result into a
            // temporary variable to avoid calling stateful or impure code more than once.
            temporary = this.allocateTemporary();
            // Preserve the result in the temporary variable
            guardedExpression = temporary.set(guardedExpression);
            // Ensure all further references to the guarded expression refer to the temporary instead.
            this._resultMap.set(leftMostSafe.receiver, temporary);
        }
        var condition = guardedExpression.isBlank();
        // Convert the ast to an unguarded access to the receiver's member. The map will substitute
        // leftMostNode with its unguarded version in the call to `this.visit()`.
        if (leftMostSafe instanceof cdAst.SafeMethodCall) {
            this._nodeMap.set(leftMostSafe, new cdAst.MethodCall(leftMostSafe.span, leftMostSafe.receiver, leftMostSafe.name, leftMostSafe.args));
        }
        else {
            this._nodeMap.set(leftMostSafe, new cdAst.PropertyRead(leftMostSafe.span, leftMostSafe.receiver, leftMostSafe.name));
        }
        // Recursively convert the node now without the guarded member access.
        var access = this.visit(ast, _Mode.Expression);
        // Remove the mapping. This is not strictly required as the converter only traverses each node
        // once but is safer if the conversion is changed to traverse the nodes more than once.
        this._nodeMap.delete(leftMostSafe);
        // If we allcoated a temporary, release it.
        if (temporary) {
            this.releaseTemporary(temporary);
        }
        // Produce the conditional
        return convertToStatementIfNeeded(mode, condition.conditional(o.literal(null), access));
    };
    // Given a expression of the form a?.b.c?.d.e the the left most safe node is
    // the (a?.b). The . and ?. are left associative thus can be rewritten as:
    // ((((a?.c).b).c)?.d).e. This returns the most deeply nested safe read or
    // safe method call as this needs be transform initially to:
    //   a == null ? null : a.c.b.c?.d.e
    // then to:
    //   a == null ? null : a.b.c == null ? null : a.b.c.d.e
    _AstToIrVisitor.prototype.leftMostSafeNode = function (ast) {
        var _this = this;
        var visit = function (visitor, ast) {
            return (_this._nodeMap.get(ast) || ast).visit(visitor);
        };
        return ast.visit({
            visitBinary: function (ast) { return null; },
            visitChain: function (ast) { return null; },
            visitConditional: function (ast) { return null; },
            visitFunctionCall: function (ast) { return null; },
            visitImplicitReceiver: function (ast) { return null; },
            visitInterpolation: function (ast) { return null; },
            visitKeyedRead: function (ast) { return visit(this, ast.obj); },
            visitKeyedWrite: function (ast) { return null; },
            visitLiteralArray: function (ast) { return null; },
            visitLiteralMap: function (ast) { return null; },
            visitLiteralPrimitive: function (ast) { return null; },
            visitMethodCall: function (ast) { return visit(this, ast.receiver); },
            visitPipe: function (ast) { return null; },
            visitPrefixNot: function (ast) { return null; },
            visitPropertyRead: function (ast) { return visit(this, ast.receiver); },
            visitPropertyWrite: function (ast) { return null; },
            visitQuote: function (ast) { return null; },
            visitSafeMethodCall: function (ast) { return visit(this, ast.receiver) || ast; },
            visitSafePropertyRead: function (ast) {
                return visit(this, ast.receiver) || ast;
            }
        });
    };
    // Returns true of the AST includes a method or a pipe indicating that, if the
    // expression is used as the target of a safe property or method access then
    // the expression should be stored into a temporary variable.
    _AstToIrVisitor.prototype.needsTemporary = function (ast) {
        var _this = this;
        var visit = function (visitor, ast) {
            return ast && (_this._nodeMap.get(ast) || ast).visit(visitor);
        };
        var visitSome = function (visitor, ast) {
            return ast.some(function (ast) { return visit(visitor, ast); });
        };
        return ast.visit({
            visitBinary: function (ast) { return visit(this, ast.left) || visit(this, ast.right); },
            visitChain: function (ast) { return false; },
            visitConditional: function (ast) {
                return visit(this, ast.condition) || visit(this, ast.trueExp) ||
                    visit(this, ast.falseExp);
            },
            visitFunctionCall: function (ast) { return true; },
            visitImplicitReceiver: function (ast) { return false; },
            visitInterpolation: function (ast) { return visitSome(this, ast.expressions); },
            visitKeyedRead: function (ast) { return false; },
            visitKeyedWrite: function (ast) { return false; },
            visitLiteralArray: function (ast) { return true; },
            visitLiteralMap: function (ast) { return true; },
            visitLiteralPrimitive: function (ast) { return false; },
            visitMethodCall: function (ast) { return true; },
            visitPipe: function (ast) { return true; },
            visitPrefixNot: function (ast) { return visit(this, ast.expression); },
            visitPropertyRead: function (ast) { return false; },
            visitPropertyWrite: function (ast) { return false; },
            visitQuote: function (ast) { return false; },
            visitSafeMethodCall: function (ast) { return true; },
            visitSafePropertyRead: function (ast) { return false; }
        });
    };
    _AstToIrVisitor.prototype.allocateTemporary = function () {
        var tempNumber = this._currentTemporary++;
        this.temporaryCount = Math.max(this._currentTemporary, this.temporaryCount);
        return new o.ReadVarExpr(temporaryName(this.bindingIndex, tempNumber));
    };
    _AstToIrVisitor.prototype.releaseTemporary = function (temporary) {
        this._currentTemporary--;
        if (temporary.name != temporaryName(this.bindingIndex, this._currentTemporary)) {
            throw new Error("Temporary " + temporary.name + " released out of order");
        }
    };
    return _AstToIrVisitor;
}());
function flattenStatements(arg, output) {
    if (isArray(arg)) {
        arg.forEach(function (entry) { return flattenStatements(entry, output); });
    }
    else {
        output.push(arg);
    }
}
//# sourceMappingURL=expression_converter.js.map