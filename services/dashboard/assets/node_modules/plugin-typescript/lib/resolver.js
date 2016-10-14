System.register(['typescript', './logger', './utils'], function(exports_1, context_1) {
    "use strict";
    var __moduleName = context_1 && context_1.id;
    var ts, logger_1, utils_1;
    var logger, Resolver;
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
            Resolver = (function () {
                function Resolver(host, resolve, lookup) {
                    this._host = host;
                    this._resolve = resolve;
                    this._lookup = lookup;
                    this._declarationFiles = [];
                }
                Resolver.prototype.resolve = function (sourceName) {
                    var _this = this;
                    var file = this._host.getSourceFile(sourceName);
                    if (!file)
                        throw new Error("file [" + sourceName + "] has not been added");
                    if (!file.pendingDependencies) {
                        var info = ts.preProcessFile(file.text, true);
                        file.isLibFile = info.isLibFile;
                        file.pendingDependencies = this.resolveDependencies(sourceName, info)
                            .then(function (mappings) {
                            var deps = Object.keys(mappings)
                                .map(function (key) { return mappings[key]; })
                                .filter(function (res) { return utils_1.isTypescript(res); });
                            var refs = _this._declarationFiles.filter(function (decl) {
                                return (decl != sourceName) && (deps.indexOf(decl) < 0);
                            });
                            var list = deps.concat(refs);
                            file.dependencies = { mappings: mappings, list: list };
                            return file.dependencies;
                        });
                    }
                    return file.pendingDependencies;
                };
                Resolver.prototype.registerDeclarationFile = function (sourceName) {
                    this._declarationFiles.push(sourceName);
                };
                Resolver.prototype.resolveDependencies = function (sourceName, info) {
                    var _this = this;
                    var resolvedReferences = info.referencedFiles
                        .map(function (ref) { return _this.resolveReference(ref.fileName, sourceName); });
                    var resolvedImports = info.importedFiles
                        .map(function (imp) { return _this.resolveImport(imp.fileName, sourceName); });
                    var resolvedExternals = info.ambientExternalModules && info.ambientExternalModules
                        .map(function (ext) { return _this.resolveImport(ext, sourceName); });
                    var refs = []
                        .concat(info.referencedFiles)
                        .concat(info.importedFiles)
                        .map(function (pre) { return pre.fileName; })
                        .concat(info.ambientExternalModules);
                    var deps = resolvedReferences.concat(resolvedImports).concat(resolvedExternals);
                    return Promise.all(deps)
                        .then(function (resolved) {
                        return refs.reduce(function (result, ref, idx) {
                            result[ref] = resolved[idx];
                            return result;
                        }, {});
                    });
                };
                Resolver.prototype.resolveReference = function (referenceName, sourceName) {
                    if ((utils_1.isAmbient(referenceName) && !this._host.options.resolveAmbientRefs) || (referenceName.indexOf("/") === -1))
                        referenceName = "./" + referenceName;
                    return this._resolve(referenceName, sourceName);
                };
                Resolver.prototype.resolveImport = function (importName, sourceName) {
                    var _this = this;
                    if (utils_1.isRelative(importName) && utils_1.isTypescriptDeclaration(sourceName) && !utils_1.isTypescriptDeclaration(importName))
                        importName = importName + ".d.ts";
                    return this._resolve(importName, sourceName)
                        .then(function (address) {
                        if (utils_1.isJavaScript(address)) {
                            return _this.lookupTyping(importName, sourceName, address)
                                .then(function (typingAddress) {
                                return typingAddress ? typingAddress : address;
                            });
                        }
                        return address;
                    });
                };
                Resolver.prototype.lookupTyping = function (importName, sourceName, address) {
                    var _this = this;
                    return this._lookup(address)
                        .then(function (metadata) {
                        if (metadata.typings === true) {
                            return utils_1.jsToDts(address);
                        }
                        else if (typeof (metadata.typings) === 'string') {
                            var packageName = importName.split('/')[0];
                            var typingsName = utils_1.isRelative(metadata.typings) ? metadata.typings.slice(2) : metadata.typings;
                            return _this._resolve(packageName + "/" + typingsName, sourceName);
                        }
                        else if (metadata.typings) {
                            throw new Error("invalid 'typings' value [" + metadata.typings + "] [" + address + "]");
                        }
                        else {
                            return undefined;
                        }
                    });
                };
                return Resolver;
            }());
            exports_1("Resolver", Resolver);
        }
    }
});
