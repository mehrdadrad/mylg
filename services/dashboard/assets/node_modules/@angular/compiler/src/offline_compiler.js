/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
import { AnimationCompiler } from './animation/animation_compiler';
import { AnimationParser } from './animation/animation_parser';
import { CompileProviderMetadata, createHostComponentMeta } from './compile_metadata';
import { Identifiers, resolveIdentifier, resolveIdentifierToken } from './identifiers';
import * as o from './output/output_ast';
import { ComponentFactoryDependency, ViewFactoryDependency } from './view_compiler/view_compiler';
export var SourceModule = (function () {
    function SourceModule(moduleUrl, source) {
        this.moduleUrl = moduleUrl;
        this.source = source;
    }
    return SourceModule;
}());
export var NgModulesSummary = (function () {
    function NgModulesSummary(ngModuleByComponent) {
        this.ngModuleByComponent = ngModuleByComponent;
    }
    return NgModulesSummary;
}());
export var OfflineCompiler = (function () {
    function OfflineCompiler(_metadataResolver, _directiveNormalizer, _templateParser, _styleCompiler, _viewCompiler, _ngModuleCompiler, _outputEmitter, _localeId, _translationFormat) {
        this._metadataResolver = _metadataResolver;
        this._directiveNormalizer = _directiveNormalizer;
        this._templateParser = _templateParser;
        this._styleCompiler = _styleCompiler;
        this._viewCompiler = _viewCompiler;
        this._ngModuleCompiler = _ngModuleCompiler;
        this._outputEmitter = _outputEmitter;
        this._localeId = _localeId;
        this._translationFormat = _translationFormat;
        this._animationParser = new AnimationParser();
        this._animationCompiler = new AnimationCompiler();
    }
    OfflineCompiler.prototype.analyzeModules = function (ngModules) {
        var _this = this;
        var ngModuleByComponent = new Map();
        ngModules.forEach(function (ngModule) {
            var ngModuleMeta = _this._metadataResolver.getNgModuleMetadata(ngModule);
            ngModuleMeta.declaredDirectives.forEach(function (dirMeta) {
                if (dirMeta.isComponent) {
                    ngModuleByComponent.set(dirMeta.type.reference, ngModuleMeta);
                }
            });
        });
        return new NgModulesSummary(ngModuleByComponent);
    };
    OfflineCompiler.prototype.clearCache = function () {
        this._directiveNormalizer.clearCache();
        this._metadataResolver.clearCache();
    };
    OfflineCompiler.prototype.compile = function (moduleUrl, ngModulesSummary, components, ngModules) {
        var _this = this;
        var fileSuffix = _splitTypescriptSuffix(moduleUrl)[1];
        var statements = [];
        var exportedVars = [];
        var outputSourceModules = [];
        // compile all ng modules
        exportedVars.push.apply(exportedVars, ngModules.map(function (ngModuleType) { return _this._compileModule(ngModuleType, statements); }));
        // compile components
        return Promise
            .all(components.map(function (compType) {
            var compMeta = _this._metadataResolver.getDirectiveMetadata(compType);
            var ngModule = ngModulesSummary.ngModuleByComponent.get(compType);
            if (!ngModule) {
                throw new Error("Cannot determine the module for component " + compMeta.type.name + "!");
            }
            return Promise
                .all([compMeta].concat(ngModule.transitiveModule.directives).map(function (dirMeta) { return _this._directiveNormalizer.normalizeDirective(dirMeta).asyncResult; }))
                .then(function (normalizedCompWithDirectives) {
                var compMeta = normalizedCompWithDirectives[0], dirMetas = normalizedCompWithDirectives.slice(1);
                _assertComponent(compMeta);
                // compile styles
                var stylesCompileResults = _this._styleCompiler.compileComponent(compMeta);
                stylesCompileResults.externalStylesheets.forEach(function (compiledStyleSheet) {
                    outputSourceModules.push(_this._codgenStyles(compiledStyleSheet, fileSuffix));
                });
                // compile components
                exportedVars.push(_this._compileComponentFactory(compMeta, fileSuffix, statements), _this._compileComponent(compMeta, dirMetas, ngModule.transitiveModule.pipes, ngModule.schemas, stylesCompileResults.componentStylesheet, fileSuffix, statements));
            });
        }))
            .then(function () {
            if (statements.length > 0) {
                outputSourceModules.unshift(_this._codegenSourceModule(_ngfactoryModuleUrl(moduleUrl), statements, exportedVars));
            }
            return outputSourceModules;
        });
    };
    OfflineCompiler.prototype._compileModule = function (ngModuleType, targetStatements) {
        var ngModule = this._metadataResolver.getNgModuleMetadata(ngModuleType);
        var providers = [];
        if (this._localeId) {
            providers.push(new CompileProviderMetadata({
                token: resolveIdentifierToken(Identifiers.LOCALE_ID),
                useValue: this._localeId,
            }));
        }
        if (this._translationFormat) {
            providers.push(new CompileProviderMetadata({
                token: resolveIdentifierToken(Identifiers.TRANSLATIONS_FORMAT),
                useValue: this._translationFormat
            }));
        }
        var appCompileResult = this._ngModuleCompiler.compile(ngModule, providers);
        appCompileResult.dependencies.forEach(function (dep) {
            dep.placeholder.name = _componentFactoryName(dep.comp);
            dep.placeholder.moduleUrl = _ngfactoryModuleUrl(dep.comp.moduleUrl);
        });
        targetStatements.push.apply(targetStatements, appCompileResult.statements);
        return appCompileResult.ngModuleFactoryVar;
    };
    OfflineCompiler.prototype._compileComponentFactory = function (compMeta, fileSuffix, targetStatements) {
        var hostMeta = createHostComponentMeta(compMeta);
        var hostViewFactoryVar = this._compileComponent(hostMeta, [compMeta], [], [], null, fileSuffix, targetStatements);
        var compFactoryVar = _componentFactoryName(compMeta.type);
        targetStatements.push(o.variable(compFactoryVar)
            .set(o.importExpr(resolveIdentifier(Identifiers.ComponentFactory), [o.importType(compMeta.type)])
            .instantiate([
            o.literal(compMeta.selector),
            o.variable(hostViewFactoryVar),
            o.importExpr(compMeta.type),
        ], o.importType(resolveIdentifier(Identifiers.ComponentFactory), [o.importType(compMeta.type)], [o.TypeModifier.Const])))
            .toDeclStmt(null, [o.StmtModifier.Final]));
        return compFactoryVar;
    };
    OfflineCompiler.prototype._compileComponent = function (compMeta, directives, pipes, schemas, componentStyles, fileSuffix, targetStatements) {
        var parsedAnimations = this._animationParser.parseComponent(compMeta);
        var parsedTemplate = this._templateParser.parse(compMeta, compMeta.template.template, directives, pipes, schemas, compMeta.type.name);
        var stylesExpr = componentStyles ? o.variable(componentStyles.stylesVar) : o.literalArr([]);
        var compiledAnimations = this._animationCompiler.compile(compMeta.type.name, parsedAnimations);
        var viewResult = this._viewCompiler.compileComponent(compMeta, parsedTemplate, stylesExpr, pipes, compiledAnimations);
        if (componentStyles) {
            targetStatements.push.apply(targetStatements, _resolveStyleStatements(componentStyles, fileSuffix));
        }
        compiledAnimations.forEach(function (entry) { entry.statements.forEach(function (statement) { targetStatements.push(statement); }); });
        targetStatements.push.apply(targetStatements, _resolveViewStatements(viewResult));
        return viewResult.viewFactoryVar;
    };
    OfflineCompiler.prototype._codgenStyles = function (stylesCompileResult, fileSuffix) {
        _resolveStyleStatements(stylesCompileResult, fileSuffix);
        return this._codegenSourceModule(_stylesModuleUrl(stylesCompileResult.meta.moduleUrl, stylesCompileResult.isShimmed, fileSuffix), stylesCompileResult.statements, [stylesCompileResult.stylesVar]);
    };
    OfflineCompiler.prototype._codegenSourceModule = function (moduleUrl, statements, exportedVars) {
        return new SourceModule(moduleUrl, this._outputEmitter.emitStatements(moduleUrl, statements, exportedVars));
    };
    return OfflineCompiler;
}());
function _resolveViewStatements(compileResult) {
    compileResult.dependencies.forEach(function (dep) {
        if (dep instanceof ViewFactoryDependency) {
            var vfd = dep;
            vfd.placeholder.moduleUrl = _ngfactoryModuleUrl(vfd.comp.moduleUrl);
        }
        else if (dep instanceof ComponentFactoryDependency) {
            var cfd = dep;
            cfd.placeholder.name = _componentFactoryName(cfd.comp);
            cfd.placeholder.moduleUrl = _ngfactoryModuleUrl(cfd.comp.moduleUrl);
        }
    });
    return compileResult.statements;
}
function _resolveStyleStatements(compileResult, fileSuffix) {
    compileResult.dependencies.forEach(function (dep) {
        dep.valuePlaceholder.moduleUrl = _stylesModuleUrl(dep.moduleUrl, dep.isShimmed, fileSuffix);
    });
    return compileResult.statements;
}
function _ngfactoryModuleUrl(compUrl) {
    var urlWithSuffix = _splitTypescriptSuffix(compUrl);
    return urlWithSuffix[0] + ".ngfactory" + urlWithSuffix[1];
}
function _componentFactoryName(comp) {
    return comp.name + "NgFactory";
}
function _stylesModuleUrl(stylesheetUrl, shim, suffix) {
    return shim ? stylesheetUrl + ".shim" + suffix : "" + stylesheetUrl + suffix;
}
function _assertComponent(meta) {
    if (!meta.isComponent) {
        throw new Error("Could not compile '" + meta.type.name + "' because it is not a component.");
    }
}
function _splitTypescriptSuffix(path) {
    if (path.endsWith('.d.ts')) {
        return [path.slice(0, -5), '.ts'];
    }
    var lastDot = path.lastIndexOf('.');
    if (lastDot !== -1) {
        return [path.substring(0, lastDot), path.substring(lastDot)];
    }
    return [path, ''];
}
//# sourceMappingURL=offline_compiler.js.map