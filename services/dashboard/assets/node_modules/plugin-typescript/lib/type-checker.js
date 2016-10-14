System.register(['typescript', './logger', './utils', "./compiler-host"], function(exports_1, context_1) {
    "use strict";
    var __moduleName = context_1 && context_1.id;
    var ts, logger_1, utils_1, compiler_host_1;
    var logger, TypeChecker;
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
            },
            function (compiler_host_1_1) {
                compiler_host_1 = compiler_host_1_1;
            }],
        execute: function() {
            logger = new logger_1.default({ debug: false });
            TypeChecker = (function () {
                function TypeChecker(host) {
                    this._host = host;
                    this._options = ts.clone(this._host.options);
                    this._options.inlineSourceMap = false;
                    this._options.sourceMap = false;
                    this._options.declaration = false;
                    this._options.isolatedModules = false;
                    this._options.skipDefaultLibCheck = true;
                }
                TypeChecker.prototype.check = function () {
                    var candidates = this.getCandidates();
                    console.log(JSON.stringify(candidates.map(function (c) { return ({ name: c.name, checkable: c.checkable }); })));
                    if (candidates.some(function (candidate) { return utils_1.isTypescriptDeclaration(candidate.name) && !candidate.checkable; }))
                        return [];
                    else if (candidates.some(function (candidate) { return candidate.checkable && !utils_1.isTypescriptDeclaration(candidate.name); }))
                        return this.getAllDiagnostics(candidates);
                    else
                        return [];
                };
                TypeChecker.prototype.forceCheck = function () {
                    var candidates = this.getCandidates();
                    candidates.forEach(function (candidate) { return candidate.checkable = true; });
                    return this.getAllDiagnostics(candidates);
                };
                TypeChecker.prototype.hasErrors = function () {
                    return this._host.getAllFiles()
                        .filter(function (file) { return file.fileName != compiler_host_1.__HTML_MODULE__; })
                        .some(function (file) { return file.checked && utils_1.hasError(file.errors); });
                };
                TypeChecker.prototype.getCandidates = function () {
                    var _this = this;
                    var candidates = this._host.getAllFiles()
                        .filter(function (file) { return file.fileName != compiler_host_1.__HTML_MODULE__; })
                        .map(function (file) { return ({
                        name: file.fileName,
                        file: file,
                        seen: false,
                        resolved: !!file.dependencies,
                        checkable: undefined,
                        deps: file.dependencies && file.dependencies.list
                    }); });
                    var candidatesMap = candidates.reduce(function (result, candidate) {
                        result[candidate.name] = candidate;
                        return result;
                    }, {});
                    candidates.forEach(function (candidate) { return candidate.checkable = _this.isCheckable(candidate, candidatesMap); });
                    return candidates;
                };
                TypeChecker.prototype.isCheckable = function (candidate, candidatesMap) {
                    var _this = this;
                    if (!candidate)
                        return false;
                    else {
                        if (!candidate.seen) {
                            candidate.seen = true;
                            candidate.checkable = candidate.resolved && candidate.deps.every(function (dep) { return _this.isCheckable(candidatesMap[dep], candidatesMap); });
                        }
                        return (candidate.checkable !== false);
                    }
                };
                TypeChecker.prototype.getAllDiagnostics = function (candidates) {
                    var filelist = candidates.map(function (dep) { return dep.name; }).concat([compiler_host_1.__HTML_MODULE__]);
                    var program = ts.createProgram(filelist, this._options, this._host);
                    return candidates.reduce(function (errors, candidate) {
                        if (candidate.checkable && !candidate.file.checked) {
                            candidate.file.errors = [];
                            if (!candidate.file.isLibFile) {
                                candidate.file.errors = program.getSyntacticDiagnostics(candidate.file)
                                    .concat(program.getSemanticDiagnostics(candidate.file));
                            }
                            candidate.file.checked = true;
                            return errors.concat(candidate.file.errors);
                        }
                        else {
                            return errors;
                        }
                    }, program.getGlobalDiagnostics());
                };
                return TypeChecker;
            }());
            exports_1("TypeChecker", TypeChecker);
        }
    }
});
