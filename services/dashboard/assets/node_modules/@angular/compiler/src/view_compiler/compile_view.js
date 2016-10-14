/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
import { CompileIdentifierMetadata } from '../compile_metadata';
import { ListWrapper, MapWrapper } from '../facade/collection';
import { isPresent } from '../facade/lang';
import { Identifiers, resolveIdentifier } from '../identifiers';
import * as o from '../output/output_ast';
import { ViewType } from '../private_import_core';
import { CompileMethod } from './compile_method';
import { CompilePipe } from './compile_pipe';
import { CompileQuery, addQueryToTokenMap, createQueryList } from './compile_query';
import { EventHandlerVars } from './constants';
import { createPureProxy, getPropertyInView, getViewFactoryName } from './util';
export var CompileView = (function () {
    function CompileView(component, genConfig, pipeMetas, styles, animations, viewIndex, declarationElement, templateVariableBindings) {
        var _this = this;
        this.component = component;
        this.genConfig = genConfig;
        this.pipeMetas = pipeMetas;
        this.styles = styles;
        this.animations = animations;
        this.viewIndex = viewIndex;
        this.declarationElement = declarationElement;
        this.templateVariableBindings = templateVariableBindings;
        this.nodes = [];
        // root nodes or AppElements for ViewContainers
        this.rootNodesOrAppElements = [];
        this.bindings = [];
        this.classStatements = [];
        this.eventHandlerMethods = [];
        this.fields = [];
        this.getters = [];
        this.disposables = [];
        this.subscriptions = [];
        this.purePipes = new Map();
        this.pipes = [];
        this.locals = new Map();
        this.literalArrayCount = 0;
        this.literalMapCount = 0;
        this.pipeCount = 0;
        this.createMethod = new CompileMethod(this);
        this.animationBindingsMethod = new CompileMethod(this);
        this.injectorGetMethod = new CompileMethod(this);
        this.updateContentQueriesMethod = new CompileMethod(this);
        this.dirtyParentQueriesMethod = new CompileMethod(this);
        this.updateViewQueriesMethod = new CompileMethod(this);
        this.detectChangesInInputsMethod = new CompileMethod(this);
        this.detectChangesRenderPropertiesMethod = new CompileMethod(this);
        this.afterContentLifecycleCallbacksMethod = new CompileMethod(this);
        this.afterViewLifecycleCallbacksMethod = new CompileMethod(this);
        this.destroyMethod = new CompileMethod(this);
        this.detachMethod = new CompileMethod(this);
        this.viewType = getViewType(component, viewIndex);
        this.className = "_View_" + component.type.name + viewIndex;
        this.classType = o.importType(new CompileIdentifierMetadata({ name: this.className }));
        this.viewFactory = o.variable(getViewFactoryName(component, viewIndex));
        if (this.viewType === ViewType.COMPONENT || this.viewType === ViewType.HOST) {
            this.componentView = this;
        }
        else {
            this.componentView = this.declarationElement.view.componentView;
        }
        this.componentContext =
            getPropertyInView(o.THIS_EXPR.prop('context'), this, this.componentView);
        var viewQueries = new Map();
        if (this.viewType === ViewType.COMPONENT) {
            var directiveInstance = o.THIS_EXPR.prop('context');
            ListWrapper.forEachWithIndex(this.component.viewQueries, function (queryMeta, queryIndex) {
                var propName = "_viewQuery_" + queryMeta.selectors[0].name + "_" + queryIndex;
                var queryList = createQueryList(queryMeta, directiveInstance, propName, _this);
                var query = new CompileQuery(queryMeta, queryList, directiveInstance, _this);
                addQueryToTokenMap(viewQueries, query);
            });
            var constructorViewQueryCount = 0;
            this.component.type.diDeps.forEach(function (dep) {
                if (isPresent(dep.viewQuery)) {
                    var queryList = o.THIS_EXPR.prop('declarationAppElement')
                        .prop('componentConstructorViewQueries')
                        .key(o.literal(constructorViewQueryCount++));
                    var query = new CompileQuery(dep.viewQuery, queryList, null, _this);
                    addQueryToTokenMap(viewQueries, query);
                }
            });
        }
        this.viewQueries = viewQueries;
        templateVariableBindings.forEach(function (entry) { _this.locals.set(entry[1], o.THIS_EXPR.prop('context').prop(entry[0])); });
        if (!this.declarationElement.isNull()) {
            this.declarationElement.setEmbeddedView(this);
        }
    }
    CompileView.prototype.callPipe = function (name, input, args) {
        return CompilePipe.call(this, name, [input].concat(args));
    };
    CompileView.prototype.getLocal = function (name) {
        if (name == EventHandlerVars.event.name) {
            return EventHandlerVars.event;
        }
        var currView = this;
        var result = currView.locals.get(name);
        while (!result && isPresent(currView.declarationElement.view)) {
            currView = currView.declarationElement.view;
            result = currView.locals.get(name);
        }
        if (isPresent(result)) {
            return getPropertyInView(result, this, currView);
        }
        else {
            return null;
        }
    };
    CompileView.prototype.createLiteralArray = function (values) {
        if (values.length === 0) {
            return o.importExpr(resolveIdentifier(Identifiers.EMPTY_ARRAY));
        }
        var proxyExpr = o.THIS_EXPR.prop("_arr_" + this.literalArrayCount++);
        var proxyParams = [];
        var proxyReturnEntries = [];
        for (var i = 0; i < values.length; i++) {
            var paramName = "p" + i;
            proxyParams.push(new o.FnParam(paramName));
            proxyReturnEntries.push(o.variable(paramName));
        }
        createPureProxy(o.fn(proxyParams, [new o.ReturnStatement(o.literalArr(proxyReturnEntries))], new o.ArrayType(o.DYNAMIC_TYPE)), values.length, proxyExpr, this);
        return proxyExpr.callFn(values);
    };
    CompileView.prototype.createLiteralMap = function (entries) {
        if (entries.length === 0) {
            return o.importExpr(resolveIdentifier(Identifiers.EMPTY_MAP));
        }
        var proxyExpr = o.THIS_EXPR.prop("_map_" + this.literalMapCount++);
        var proxyParams = [];
        var proxyReturnEntries = [];
        var values = [];
        for (var i = 0; i < entries.length; i++) {
            var paramName = "p" + i;
            proxyParams.push(new o.FnParam(paramName));
            proxyReturnEntries.push([entries[i][0], o.variable(paramName)]);
            values.push(entries[i][1]);
        }
        createPureProxy(o.fn(proxyParams, [new o.ReturnStatement(o.literalMap(proxyReturnEntries))], new o.MapType(o.DYNAMIC_TYPE)), entries.length, proxyExpr, this);
        return proxyExpr.callFn(values);
    };
    CompileView.prototype.afterNodes = function () {
        var _this = this;
        MapWrapper.values(this.viewQueries)
            .forEach(function (queries) { return queries.forEach(function (query) { return query.afterChildren(_this.createMethod, _this.updateViewQueriesMethod); }); });
    };
    return CompileView;
}());
function getViewType(component, embeddedTemplateIndex) {
    if (embeddedTemplateIndex > 0) {
        return ViewType.EMBEDDED;
    }
    else if (component.type.isHost) {
        return ViewType.HOST;
    }
    else {
        return ViewType.COMPONENT;
    }
}
//# sourceMappingURL=compile_view.js.map