System.register(['typescript', './logger', './compiler-host', './transpiler', './resolver', './type-checker', './format-errors', './utils'], function(exports_1, context_1) {
    "use strict";
    var __moduleName = context_1 && context_1.id;
    var ts, logger_1, compiler_host_1, transpiler_1, resolver_1, type_checker_1, format_errors_1, utils_1;
    var logger;
    function createFactory(sjsconfig, builder, _resolve, _fetch, _lookup) {
        if (sjsconfig === void 0) { sjsconfig = {}; }
        var tsconfigFiles = [];
        var typingsFiles = [];
        return loadOptions(sjsconfig, _resolve, _fetch)
            .then(function (options) {
            return createServices(options, builder, _resolve, _lookup);
        })
            .then(function (services) {
            if (services.options.typeCheck) {
                return resolveDeclarationFiles(services.options, _resolve)
                    .then(function (resolvedFiles) {
                    resolvedFiles.forEach(function (resolvedFile) {
                        services.resolver.registerDeclarationFile(resolvedFile);
                    });
                    return services;
                });
            }
            else {
                return services;
            }
        });
    }
    exports_1("createFactory", createFactory);
    function loadOptions(sjsconfig, _resolve, _fetch) {
        if (sjsconfig.tsconfig) {
            var tsconfig = (sjsconfig.tsconfig === true) ? "tsconfig.json" : sjsconfig.tsconfig;
            return _resolve(tsconfig)
                .then(function (tsconfigAddress) {
                return _fetch(tsconfigAddress).then(function (tsconfigText) { return ({ tsconfigText: tsconfigText, tsconfigAddress: tsconfigAddress }); });
            })
                .then(function (_a) {
                var tsconfigAddress = _a.tsconfigAddress, tsconfigText = _a.tsconfigText;
                var result = ts.parseConfigFileTextToJson(tsconfigAddress, tsconfigText);
                if (result.error) {
                    format_errors_1.formatErrors([result.error], logger);
                    throw new Error("failed to load tsconfig from " + tsconfigAddress);
                }
                var files = result.config.files;
                return ts.extend(ts.extend({ tsconfigAddress: tsconfigAddress, files: files }, sjsconfig), result.config.compilerOptions);
            });
        }
        else {
            return Promise.resolve(sjsconfig);
        }
    }
    function resolveDeclarationFiles(options, _resolve) {
        var files = options.files || [];
        var declarationFiles = files
            .filter(function (filename) { return utils_1.isTypescriptDeclaration(filename); })
            .map(function (filename) { return _resolve(filename, options.tsconfigAddress); });
        return Promise.all(declarationFiles);
    }
    function createServices(options, builder, _resolve, _lookup) {
        var host = new compiler_host_1.CompilerHost(options, builder);
        var transpiler = new transpiler_1.Transpiler(host);
        var resolver = undefined;
        var typeChecker = undefined;
        if (options.typeCheck) {
            resolver = new resolver_1.Resolver(host, _resolve, _lookup);
            typeChecker = new type_checker_1.TypeChecker(host);
            if (!host.options.noLib) {
                return _resolve(host.getDefaultLibFileName())
                    .then(function (defaultLibAddress) {
                    resolver.registerDeclarationFile(defaultLibAddress);
                    return { host: host, transpiler: transpiler, resolver: resolver, typeChecker: typeChecker, options: options };
                });
            }
        }
        return Promise.resolve({ host: host, transpiler: transpiler, resolver: resolver, typeChecker: typeChecker, options: options });
    }
    return {
        setters:[
            function (ts_1) {
                ts = ts_1;
            },
            function (logger_1_1) {
                logger_1 = logger_1_1;
            },
            function (compiler_host_1_1) {
                compiler_host_1 = compiler_host_1_1;
            },
            function (transpiler_1_1) {
                transpiler_1 = transpiler_1_1;
            },
            function (resolver_1_1) {
                resolver_1 = resolver_1_1;
            },
            function (type_checker_1_1) {
                type_checker_1 = type_checker_1_1;
            },
            function (format_errors_1_1) {
                format_errors_1 = format_errors_1_1;
            },
            function (utils_1_1) {
                utils_1 = utils_1_1;
            }],
        execute: function() {
            logger = new logger_1.default({ debug: false });
        }
    }
});
