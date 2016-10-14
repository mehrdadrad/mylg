System.register(['typescript', './utils', './logger'], function(exports_1, context_1) {
    "use strict";
    var __moduleName = context_1 && context_1.id;
    var ts, utils_1, logger_1;
    var logger, Transpiler;
    return {
        setters:[
            function (ts_1) {
                ts = ts_1;
            },
            function (utils_1_1) {
                utils_1 = utils_1_1;
            },
            function (logger_1_1) {
                logger_1 = logger_1_1;
            }],
        execute: function() {
            logger = new logger_1.default({ debug: false });
            Transpiler = (function () {
                function Transpiler(host) {
                    this._host = host;
                    this._options = ts.clone(this._host.options);
                    this._options.isolatedModules = true;
                    if (this._options.sourceMap === undefined)
                        this._options.sourceMap = this._options.inlineSourceMap;
                    if (this._options.sourceMap === undefined)
                        this._options.sourceMap = true;
                    this._options.inlineSourceMap = false;
                    this._options.declaration = false;
                    this._options.noEmitOnError = false;
                    this._options.out = undefined;
                    this._options.outFile = undefined;
                    this._options.noLib = true;
                    this._options.suppressOutputPathCheck = true;
                    this._options.noEmit = false;
                }
                Transpiler.prototype.transpile = function (sourceName) {
                    logger.debug("transpiling " + sourceName);
                    var file = this._host.getSourceFile(sourceName);
                    if (!file)
                        throw new Error("file [" + sourceName + "] has not been added");
                    if (!file.output) {
                        var program = ts.createProgram([sourceName], this._options, this._host);
                        var jstext_1 = undefined;
                        var maptext_1 = undefined;
                        var emitResult = program.emit(undefined, function (outputName, output) {
                            if (utils_1.isJavaScript(outputName))
                                jstext_1 = output.slice(0, output.lastIndexOf("//#"));
                            else if (utils_1.isSourceMap(outputName))
                                maptext_1 = output;
                            else
                                throw new Error("unexpected ouput file " + outputName);
                        });
                        var diagnostics = emitResult.diagnostics
                            .concat(program.getOptionsDiagnostics())
                            .concat(program.getSyntacticDiagnostics());
                        file.output = {
                            failure: utils_1.hasError(diagnostics),
                            errors: diagnostics,
                            js: jstext_1,
                            sourceMap: maptext_1
                        };
                    }
                    return file.output;
                };
                return Transpiler;
            }());
            exports_1("Transpiler", Transpiler);
        }
    }
});
