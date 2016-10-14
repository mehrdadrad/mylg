/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
import { SecurityContext } from '@angular/core';
import { isPresent } from '../facade/lang';
import { Identifiers, resolveIdentifier } from '../identifiers';
import * as o from '../output/output_ast';
import { EMPTY_STATE as EMPTY_ANIMATION_STATE, LifecycleHooks, isDefaultChangeDetectionStrategy } from '../private_import_core';
import { PropertyBindingType } from '../template_parser/template_ast';
import { camelCaseToDashCase } from '../util';
import { CompileBinding } from './compile_binding';
import { DetectChangesVars, ViewProperties } from './constants';
import { convertCdExpressionToIr, temporaryDeclaration } from './expression_converter';
function createBindFieldExpr(exprIndex) {
    return o.THIS_EXPR.prop("_expr_" + exprIndex);
}
function createCurrValueExpr(exprIndex) {
    return o.variable("currVal_" + exprIndex); // fix syntax highlighting: `
}
function bind(view, currValExpr, fieldExpr, parsedExpression, context, actions, method, bindingIndex) {
    var checkExpression = convertCdExpressionToIr(view, context, parsedExpression, DetectChangesVars.valUnwrapper, bindingIndex);
    if (!checkExpression.expression) {
        // e.g. an empty expression was given
        return;
    }
    if (checkExpression.temporaryCount) {
        for (var i = 0; i < checkExpression.temporaryCount; i++) {
            method.addStmt(temporaryDeclaration(bindingIndex, i));
        }
    }
    // private is fine here as no child view will reference the cached value...
    view.fields.push(new o.ClassField(fieldExpr.name, null, [o.StmtModifier.Private]));
    view.createMethod.addStmt(o.THIS_EXPR.prop(fieldExpr.name)
        .set(o.importExpr(resolveIdentifier(Identifiers.UNINITIALIZED)))
        .toStmt());
    if (checkExpression.needsValueUnwrapper) {
        var initValueUnwrapperStmt = DetectChangesVars.valUnwrapper.callMethod('reset', []).toStmt();
        method.addStmt(initValueUnwrapperStmt);
    }
    method.addStmt(currValExpr.set(checkExpression.expression).toDeclStmt(null, [o.StmtModifier.Final]));
    var condition = o.importExpr(resolveIdentifier(Identifiers.checkBinding)).callFn([
        DetectChangesVars.throwOnChange, fieldExpr, currValExpr
    ]);
    if (checkExpression.needsValueUnwrapper) {
        condition = DetectChangesVars.valUnwrapper.prop('hasWrappedValue').or(condition);
    }
    method.addStmt(new o.IfStmt(condition, actions.concat([o.THIS_EXPR.prop(fieldExpr.name).set(currValExpr).toStmt()])));
}
export function bindRenderText(boundText, compileNode, view) {
    var bindingIndex = view.bindings.length;
    view.bindings.push(new CompileBinding(compileNode, boundText));
    var currValExpr = createCurrValueExpr(bindingIndex);
    var valueField = createBindFieldExpr(bindingIndex);
    view.detectChangesRenderPropertiesMethod.resetDebugInfo(compileNode.nodeIndex, boundText);
    bind(view, currValExpr, valueField, boundText.value, view.componentContext, [o.THIS_EXPR.prop('renderer')
            .callMethod('setText', [compileNode.renderNode, currValExpr])
            .toStmt()], view.detectChangesRenderPropertiesMethod, bindingIndex);
}
function bindAndWriteToRenderer(boundProps, context, compileElement, isHostProp) {
    var view = compileElement.view;
    var renderNode = compileElement.renderNode;
    boundProps.forEach(function (boundProp) {
        var bindingIndex = view.bindings.length;
        view.bindings.push(new CompileBinding(compileElement, boundProp));
        view.detectChangesRenderPropertiesMethod.resetDebugInfo(compileElement.nodeIndex, boundProp);
        var fieldExpr = createBindFieldExpr(bindingIndex);
        var currValExpr = createCurrValueExpr(bindingIndex);
        var renderMethod;
        var oldRenderValue = sanitizedValue(boundProp, fieldExpr);
        var renderValue = sanitizedValue(boundProp, currValExpr);
        var updateStmts = [];
        var compileMethod = view.detectChangesRenderPropertiesMethod;
        switch (boundProp.type) {
            case PropertyBindingType.Property:
                if (view.genConfig.logBindingUpdate) {
                    updateStmts.push(logBindingUpdateStmt(renderNode, boundProp.name, renderValue));
                }
                updateStmts.push(o.THIS_EXPR.prop('renderer')
                    .callMethod('setElementProperty', [renderNode, o.literal(boundProp.name), renderValue])
                    .toStmt());
                break;
            case PropertyBindingType.Attribute:
                renderValue =
                    renderValue.isBlank().conditional(o.NULL_EXPR, renderValue.callMethod('toString', []));
                updateStmts.push(o.THIS_EXPR.prop('renderer')
                    .callMethod('setElementAttribute', [renderNode, o.literal(boundProp.name), renderValue])
                    .toStmt());
                break;
            case PropertyBindingType.Class:
                updateStmts.push(o.THIS_EXPR.prop('renderer')
                    .callMethod('setElementClass', [renderNode, o.literal(boundProp.name), renderValue])
                    .toStmt());
                break;
            case PropertyBindingType.Style:
                var strValue = renderValue.callMethod('toString', []);
                if (isPresent(boundProp.unit)) {
                    strValue = strValue.plus(o.literal(boundProp.unit));
                }
                renderValue = renderValue.isBlank().conditional(o.NULL_EXPR, strValue);
                updateStmts.push(o.THIS_EXPR.prop('renderer')
                    .callMethod('setElementStyle', [renderNode, o.literal(boundProp.name), renderValue])
                    .toStmt());
                break;
            case PropertyBindingType.Animation:
                var animationName = boundProp.name;
                var targetViewExpr = o.THIS_EXPR;
                if (isHostProp) {
                    targetViewExpr = compileElement.appElement.prop('componentView');
                }
                compileMethod = view.animationBindingsMethod;
                var animationFnExpr = targetViewExpr.prop('componentType').prop('animations').key(o.literal(animationName));
                // it's important to normalize the void value as `void` explicitly
                // so that the styles data can be obtained from the stringmap
                var emptyStateValue = o.literal(EMPTY_ANIMATION_STATE);
                // void => ...
                var oldRenderVar = o.variable('oldRenderVar');
                updateStmts.push(oldRenderVar.set(oldRenderValue).toDeclStmt());
                updateStmts.push(new o.IfStmt(oldRenderVar.equals(o.importExpr(resolveIdentifier(Identifiers.UNINITIALIZED))), [oldRenderVar.set(emptyStateValue).toStmt()]));
                // ... => void
                var newRenderVar = o.variable('newRenderVar');
                updateStmts.push(newRenderVar.set(renderValue).toDeclStmt());
                updateStmts.push(new o.IfStmt(newRenderVar.equals(o.importExpr(resolveIdentifier(Identifiers.UNINITIALIZED))), [newRenderVar.set(emptyStateValue).toStmt()]));
                updateStmts.push(animationFnExpr.callFn([o.THIS_EXPR, renderNode, oldRenderVar, newRenderVar]).toStmt());
                view.detachMethod.addStmt(animationFnExpr.callFn([o.THIS_EXPR, renderNode, oldRenderValue, emptyStateValue])
                    .toStmt());
                break;
        }
        bind(view, currValExpr, fieldExpr, boundProp.value, context, updateStmts, compileMethod, view.bindings.length);
    });
}
function sanitizedValue(boundProp, renderValue) {
    var enumValue;
    switch (boundProp.securityContext) {
        case SecurityContext.NONE:
            return renderValue; // No sanitization needed.
        case SecurityContext.HTML:
            enumValue = 'HTML';
            break;
        case SecurityContext.STYLE:
            enumValue = 'STYLE';
            break;
        case SecurityContext.SCRIPT:
            enumValue = 'SCRIPT';
            break;
        case SecurityContext.URL:
            enumValue = 'URL';
            break;
        case SecurityContext.RESOURCE_URL:
            enumValue = 'RESOURCE_URL';
            break;
        default:
            throw new Error("internal error, unexpected SecurityContext " + boundProp.securityContext + ".");
    }
    var ctx = ViewProperties.viewUtils.prop('sanitizer');
    var args = [o.importExpr(resolveIdentifier(Identifiers.SecurityContext)).prop(enumValue), renderValue];
    return ctx.callMethod('sanitize', args);
}
export function bindRenderInputs(boundProps, compileElement) {
    bindAndWriteToRenderer(boundProps, compileElement.view.componentContext, compileElement, false);
}
export function bindDirectiveHostProps(directiveAst, directiveInstance, compileElement) {
    bindAndWriteToRenderer(directiveAst.hostProperties, directiveInstance, compileElement, true);
}
export function bindDirectiveInputs(directiveAst, directiveInstance, compileElement) {
    if (directiveAst.inputs.length === 0) {
        return;
    }
    var view = compileElement.view;
    var detectChangesInInputsMethod = view.detectChangesInInputsMethod;
    detectChangesInInputsMethod.resetDebugInfo(compileElement.nodeIndex, compileElement.sourceAst);
    var lifecycleHooks = directiveAst.directive.type.lifecycleHooks;
    var calcChangesMap = lifecycleHooks.indexOf(LifecycleHooks.OnChanges) !== -1;
    var isOnPushComp = directiveAst.directive.isComponent &&
        !isDefaultChangeDetectionStrategy(directiveAst.directive.changeDetection);
    if (calcChangesMap) {
        detectChangesInInputsMethod.addStmt(DetectChangesVars.changes.set(o.NULL_EXPR).toStmt());
    }
    if (isOnPushComp) {
        detectChangesInInputsMethod.addStmt(DetectChangesVars.changed.set(o.literal(false)).toStmt());
    }
    directiveAst.inputs.forEach(function (input) {
        var bindingIndex = view.bindings.length;
        view.bindings.push(new CompileBinding(compileElement, input));
        detectChangesInInputsMethod.resetDebugInfo(compileElement.nodeIndex, input);
        var fieldExpr = createBindFieldExpr(bindingIndex);
        var currValExpr = createCurrValueExpr(bindingIndex);
        var statements = [directiveInstance.prop(input.directiveName).set(currValExpr).toStmt()];
        if (calcChangesMap) {
            statements.push(new o.IfStmt(DetectChangesVars.changes.identical(o.NULL_EXPR), [DetectChangesVars.changes
                    .set(o.literalMap([], new o.MapType(o.importType(resolveIdentifier(Identifiers.SimpleChange)))))
                    .toStmt()]));
            statements.push(DetectChangesVars.changes.key(o.literal(input.directiveName))
                .set(o.importExpr(resolveIdentifier(Identifiers.SimpleChange))
                .instantiate([fieldExpr, currValExpr]))
                .toStmt());
        }
        if (isOnPushComp) {
            statements.push(DetectChangesVars.changed.set(o.literal(true)).toStmt());
        }
        if (view.genConfig.logBindingUpdate) {
            statements.push(logBindingUpdateStmt(compileElement.renderNode, input.directiveName, currValExpr));
        }
        bind(view, currValExpr, fieldExpr, input.value, view.componentContext, statements, detectChangesInInputsMethod, bindingIndex);
    });
    if (isOnPushComp) {
        detectChangesInInputsMethod.addStmt(new o.IfStmt(DetectChangesVars.changed, [
            compileElement.appElement.prop('componentView').callMethod('markAsCheckOnce', []).toStmt()
        ]));
    }
}
function logBindingUpdateStmt(renderNode, propName, value) {
    var tryStmt = o.THIS_EXPR.prop('renderer')
        .callMethod('setBindingDebugInfo', [
        renderNode, o.literal("ng-reflect-" + camelCaseToDashCase(propName)),
        value.isBlank().conditional(o.NULL_EXPR, value.callMethod('toString', []))
    ])
        .toStmt();
    var catchStmt = o.THIS_EXPR.prop('renderer')
        .callMethod('setBindingDebugInfo', [
        renderNode, o.literal("ng-reflect-" + camelCaseToDashCase(propName)),
        o.literal('[ERROR] Exception while trying to serialize the value')
    ])
        .toStmt();
    return new o.TryCatchStmt([tryStmt], [catchStmt]);
}
//# sourceMappingURL=property_binder.js.map