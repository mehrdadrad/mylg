/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
import { StringWrapper, isPresent } from '../facade/lang';
import { identifierToken } from '../identifiers';
import * as o from '../output/output_ast';
import { CompileBinding } from './compile_binding';
import { CompileMethod } from './compile_method';
import { EventHandlerVars, ViewProperties } from './constants';
import { convertCdStatementToIr } from './expression_converter';
export var CompileEventListener = (function () {
    function CompileEventListener(compileElement, eventTarget, eventName, eventPhase, listenerIndex) {
        this.compileElement = compileElement;
        this.eventTarget = eventTarget;
        this.eventName = eventName;
        this.eventPhase = eventPhase;
        this._hasComponentHostListener = false;
        this._actionResultExprs = [];
        this._method = new CompileMethod(compileElement.view);
        this._methodName =
            "_handle_" + santitizeEventName(eventName) + "_" + compileElement.nodeIndex + "_" + listenerIndex;
        this._eventParam = new o.FnParam(EventHandlerVars.event.name, o.importType(this.compileElement.view.genConfig.renderTypes.renderEvent));
    }
    CompileEventListener.getOrCreate = function (compileElement, eventTarget, eventName, eventPhase, targetEventListeners) {
        var listener = targetEventListeners.find(function (listener) { return listener.eventTarget == eventTarget && listener.eventName == eventName &&
            listener.eventPhase == eventPhase; });
        if (!listener) {
            listener = new CompileEventListener(compileElement, eventTarget, eventName, eventPhase, targetEventListeners.length);
            targetEventListeners.push(listener);
        }
        return listener;
    };
    Object.defineProperty(CompileEventListener.prototype, "methodName", {
        get: function () { return this._methodName; },
        enumerable: true,
        configurable: true
    });
    CompileEventListener.prototype.addAction = function (hostEvent, directive, directiveInstance) {
        if (isPresent(directive) && directive.isComponent) {
            this._hasComponentHostListener = true;
        }
        this._method.resetDebugInfo(this.compileElement.nodeIndex, hostEvent);
        var context = isPresent(directiveInstance) ? directiveInstance :
            this.compileElement.view.componentContext;
        var actionStmts = convertCdStatementToIr(this.compileElement.view, context, hostEvent.handler, this.compileElement.nodeIndex);
        var lastIndex = actionStmts.length - 1;
        if (lastIndex >= 0) {
            var lastStatement = actionStmts[lastIndex];
            var returnExpr = convertStmtIntoExpression(lastStatement);
            var preventDefaultVar = o.variable("pd_" + this._actionResultExprs.length);
            this._actionResultExprs.push(preventDefaultVar);
            if (isPresent(returnExpr)) {
                // Note: We need to cast the result of the method call to dynamic,
                // as it might be a void method!
                actionStmts[lastIndex] =
                    preventDefaultVar.set(returnExpr.cast(o.DYNAMIC_TYPE).notIdentical(o.literal(false)))
                        .toDeclStmt(null, [o.StmtModifier.Final]);
            }
        }
        this._method.addStmts(actionStmts);
    };
    CompileEventListener.prototype.finishMethod = function () {
        var markPathToRootStart = this._hasComponentHostListener ?
            this.compileElement.appElement.prop('componentView') :
            o.THIS_EXPR;
        var resultExpr = o.literal(true);
        this._actionResultExprs.forEach(function (expr) { resultExpr = resultExpr.and(expr); });
        var stmts = [markPathToRootStart.callMethod('markPathToRootAsCheckOnce', []).toStmt()]
            .concat(this._method.finish())
            .concat([new o.ReturnStatement(resultExpr)]);
        // private is fine here as no child view will reference the event handler...
        this.compileElement.view.eventHandlerMethods.push(new o.ClassMethod(this._methodName, [this._eventParam], stmts, o.BOOL_TYPE, [o.StmtModifier.Private]));
    };
    CompileEventListener.prototype.listenToRenderer = function () {
        var listenExpr;
        var eventListener = o.THIS_EXPR.callMethod('eventHandler', [o.THIS_EXPR.prop(this._methodName).callMethod(o.BuiltinMethod.Bind, [o.THIS_EXPR])]);
        if (isPresent(this.eventTarget)) {
            listenExpr = ViewProperties.renderer.callMethod('listenGlobal', [o.literal(this.eventTarget), o.literal(this.eventName), eventListener]);
        }
        else {
            listenExpr = ViewProperties.renderer.callMethod('listen', [this.compileElement.renderNode, o.literal(this.eventName), eventListener]);
        }
        var disposable = o.variable("disposable_" + this.compileElement.view.disposables.length);
        this.compileElement.view.disposables.push(disposable);
        // private is fine here as no child view will reference the event handler...
        this.compileElement.view.createMethod.addStmt(disposable.set(listenExpr).toDeclStmt(o.FUNCTION_TYPE, [o.StmtModifier.Private]));
    };
    CompileEventListener.prototype.listenToAnimation = function () {
        var outputListener = o.THIS_EXPR.callMethod('eventHandler', [o.THIS_EXPR.prop(this._methodName).callMethod(o.BuiltinMethod.Bind, [o.THIS_EXPR])]);
        // tie the property callback method to the view animations map
        var stmt = o.THIS_EXPR
            .callMethod('registerAnimationOutput', [
            this.compileElement.renderNode, o.literal(this.eventName),
            o.literal(this.eventPhase), outputListener
        ])
            .toStmt();
        this.compileElement.view.createMethod.addStmt(stmt);
    };
    CompileEventListener.prototype.listenToDirective = function (directiveInstance, observablePropName) {
        var subscription = o.variable("subscription_" + this.compileElement.view.subscriptions.length);
        this.compileElement.view.subscriptions.push(subscription);
        var eventListener = o.THIS_EXPR.callMethod('eventHandler', [o.THIS_EXPR.prop(this._methodName).callMethod(o.BuiltinMethod.Bind, [o.THIS_EXPR])]);
        this.compileElement.view.createMethod.addStmt(subscription
            .set(directiveInstance.prop(observablePropName)
            .callMethod(o.BuiltinMethod.SubscribeObservable, [eventListener]))
            .toDeclStmt(null, [o.StmtModifier.Final]));
    };
    return CompileEventListener;
}());
export function collectEventListeners(hostEvents, dirs, compileElement) {
    var eventListeners = [];
    hostEvents.forEach(function (hostEvent) {
        compileElement.view.bindings.push(new CompileBinding(compileElement, hostEvent));
        var listener = CompileEventListener.getOrCreate(compileElement, hostEvent.target, hostEvent.name, hostEvent.phase, eventListeners);
        listener.addAction(hostEvent, null, null);
    });
    dirs.forEach(function (directiveAst) {
        var directiveInstance = compileElement.instances.get(identifierToken(directiveAst.directive.type).reference);
        directiveAst.hostEvents.forEach(function (hostEvent) {
            compileElement.view.bindings.push(new CompileBinding(compileElement, hostEvent));
            var listener = CompileEventListener.getOrCreate(compileElement, hostEvent.target, hostEvent.name, hostEvent.phase, eventListeners);
            listener.addAction(hostEvent, directiveAst.directive, directiveInstance);
        });
    });
    eventListeners.forEach(function (listener) { return listener.finishMethod(); });
    return eventListeners;
}
export function bindDirectiveOutputs(directiveAst, directiveInstance, eventListeners) {
    Object.keys(directiveAst.directive.outputs).forEach(function (observablePropName) {
        var eventName = directiveAst.directive.outputs[observablePropName];
        eventListeners.filter(function (listener) { return listener.eventName == eventName; }).forEach(function (listener) {
            listener.listenToDirective(directiveInstance, observablePropName);
        });
    });
}
export function bindRenderOutputs(eventListeners) {
    eventListeners.forEach(function (listener) {
        if (listener.eventPhase) {
            listener.listenToAnimation();
        }
        else {
            listener.listenToRenderer();
        }
    });
}
function convertStmtIntoExpression(stmt) {
    if (stmt instanceof o.ExpressionStatement) {
        return stmt.expr;
    }
    else if (stmt instanceof o.ReturnStatement) {
        return stmt.value;
    }
    return null;
}
function santitizeEventName(name) {
    return StringWrapper.replaceAll(name, /[^a-zA-Z_]/g, '_');
}
//# sourceMappingURL=event_binder.js.map