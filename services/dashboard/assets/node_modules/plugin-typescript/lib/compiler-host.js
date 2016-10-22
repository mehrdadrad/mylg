System.register(['typescript', './logger', './utils'], function(exports_1, context_1) {
    "use strict";
    var __moduleName = context_1 && context_1.id;
    var ts, logger_1, utils_1;
    var logger, __HTML_MODULE__, CompilerHost;
    return {
        setters:[
            function (ts_1) {
                ts = ts_1;
            },
            function (logger_1_1) {
                logger_1 = logger_1_1;
            },
            function (utils_1_1) {
                utils_1 = utils_1_1;
            }],
        execute: function() {
            logger = new logger_1.default({ debug: false });
            exports_1("__HTML_MODULE__", __HTML_MODULE__ = "__html_module__");
            CompilerHost = (function () {
                function CompilerHost(options, builder) {
                    if (builder === void 0) { builder = false; }
                    this._options = options || {};
                    this._options.module = this.getEnum(this._options.module, ts.ModuleKind, ts.ModuleKind.System);
                    this._options.target = this.getEnum(this._options.target, ts.ScriptTarget, ts.ScriptTarget.ES5);
                    this._options.targetLib = this.getEnum(this._options.targetLib, ts.ScriptTarget, ts.ScriptTarget.ES6);
                    this._options.jsx = this.getEnum(this._options.jsx, ts.JsxEmit, ts.JsxEmit.None);
                    this._options.allowNonTsExtensions = (this._options.allowNonTsExtensions !== false);
                    this._options.skipDefaultLibCheck = (this._options.skipDefaultLibCheck !== false);
                    this._options.supportHtmlImports = (options.supportHtmlImports !== false);
                    this._options.noResolve = true;
                    this._options.moduleResolution = ts.ModuleResolutionKind.Classic;
                    if (builder) {
                        if ((this._options.module === ts.ModuleKind.System) && (this._options.target === ts.ScriptTarget.ES6)) {
                            this._options.module = ts.ModuleKind.ES6;
                        }
                    }
                    this._files = {};
                    var source = "var __html__: string = ''; export default __html__;";
                    if (this._options.module != ts.ModuleKind.ES6)
                        source = source + "export = __html__;";
                    var file = this.addFile(__HTML_MODULE__, source);
                    file.dependencies = { list: [], mappings: {} };
                    file.checked = true;
                    file.errors = [];
                }
                CompilerHost.prototype.getEnum = function (enumValue, enumType, defaultValue) {
                    if (enumValue == undefined)
                        return defaultValue;
                    for (var enumProp in enumType) {
                        if (enumProp.toLowerCase() === enumValue.toString().toLowerCase()) {
                            if (typeof enumType[enumProp] === "string")
                                return enumType[enumType[enumProp]];
                            else
                                return enumType[enumProp];
                        }
                    }
                    throw new Error("Unrecognised value [" + enumValue + "]");
                };
                Object.defineProperty(CompilerHost.prototype, "options", {
                    get: function () {
                        return this._options;
                    },
                    enumerable: true,
                    configurable: true
                });
                CompilerHost.prototype.getDefaultLibFileName = function () {
                    if (this._options.targetLib === ts.ScriptTarget.ES6)
                        return "typescript/lib/lib.es6.d.ts";
                    else
                        return "typescript/lib/lib.d.ts";
                };
                CompilerHost.prototype.useCaseSensitiveFileNames = function () {
                    return false;
                };
                CompilerHost.prototype.getCanonicalFileName = function (fileName) {
                    return ts.normalizePath(fileName);
                };
                CompilerHost.prototype.getCurrentDirectory = function () {
                    return "";
                };
                CompilerHost.prototype.getNewLine = function () {
                    return "\n";
                };
                CompilerHost.prototype.readFile = function (fileName) {
                    throw new Error("Not implemented");
                };
                CompilerHost.prototype.writeFile = function (name, text, writeByteOrderMark) {
                    throw new Error("Not implemented");
                };
                CompilerHost.prototype.getSourceFile = function (fileName) {
                    fileName = this.getCanonicalFileName(fileName);
                    return this._files[fileName];
                };
                CompilerHost.prototype.getAllFiles = function () {
                    var _this = this;
                    return Object.keys(this._files).map(function (key) { return _this._files[key]; });
                };
                CompilerHost.prototype.fileExists = function (fileName) {
                    return !!this.getSourceFile(fileName);
                };
                CompilerHost.prototype.addFile = function (fileName, text) {
                    fileName = this.getCanonicalFileName(fileName);
                    var file = this._files[fileName];
                    if (!file) {
                        this._files[fileName] = ts.createSourceFile(fileName, text, this._options.target);
                        logger.debug("added " + fileName);
                    }
                    else if (file.text != text) {
                        this._files[fileName] = ts.createSourceFile(fileName, text, this._options.target);
                        this.invalidate(fileName);
                        logger.debug("updated " + fileName);
                    }
                    return this._files[fileName];
                };
                CompilerHost.prototype.invalidate = function (fileName, seen) {
                    var _this = this;
                    seen = seen || [];
                    if (seen.indexOf(fileName) < 0) {
                        seen.push(fileName);
                        var file = this._files[fileName];
                        if (file) {
                            file.checked = false;
                            file.errors = [];
                        }
                        Object.keys(this._files)
                            .map(function (key) { return _this._files[key]; })
                            .forEach(function (file) {
                            if (file.dependencies && file.dependencies.list.indexOf(fileName) >= 0) {
                                _this.invalidate(file.fileName, seen);
                            }
                        });
                    }
                };
                CompilerHost.prototype.resolveModuleNames = function (moduleNames, containingFile) {
                    var _this = this;
                    return moduleNames.map(function (modName) {
                        var dependencies = _this._files[containingFile].dependencies;
                        if (utils_1.isHtml(modName) && _this._options.supportHtmlImports) {
                            return { resolvedFileName: __HTML_MODULE__ };
                        }
                        else if (dependencies) {
                            var resolvedFileName = dependencies.mappings[modName];
                            var isExternalLibraryImport = utils_1.isTypescriptDeclaration(resolvedFileName);
                            return { resolvedFileName: resolvedFileName, isExternalLibraryImport: isExternalLibraryImport };
                        }
                        else {
                            return ts.resolveModuleName(modName, containingFile, _this._options, _this).resolvedModule;
                        }
                    });
                };
                return CompilerHost;
            }());
            exports_1("CompilerHost", CompilerHost);
        }
    }
});
