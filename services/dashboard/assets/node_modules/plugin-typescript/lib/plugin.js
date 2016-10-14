System.register(['typescript', './logger', './factory', './format-errors', './utils'], function(exports_1, context_1) {
    "use strict";
    var __moduleName = context_1 && context_1.id;
    var ts, logger_1, factory_1, format_errors_1, utils_1;
    var logger, factory;
    function translate(load) {
        logger.debug("systemjs translating " + load.address);
        factory = factory || factory_1.createFactory(System.typescriptOptions, this.builder, _resolve, _fetch, _lookup)
            .then(function (output) {
            validateOptions(output.host.options);
            return output;
        });
        return factory.then(function (_a) {
            var transpiler = _a.transpiler, resolver = _a.resolver, typeChecker = _a.typeChecker, host = _a.host;
            host.addFile(load.address, load.source);
            if (utils_1.isTypescriptDeclaration(load.address)) {
                load.source = "";
                load.metadata.format = 'cjs';
            }
            else {
                var result = transpiler.transpile(load.address);
                format_errors_1.formatErrors(result.errors, logger);
                if (result.failure)
                    throw new Error("TypeScript transpilation failed");
                load.source = result.js;
                if (result.sourceMap)
                    load.metadata.sourceMap = JSON.parse(result.sourceMap);
                if (host.options.module === ts.ModuleKind.System)
                    load.metadata.format = 'register';
                else if (host.options.module === ts.ModuleKind.ES6)
                    load.metadata.format = 'esm';
                else if (host.options.module === ts.ModuleKind.CommonJS)
                    load.metadata.format = 'cjs';
            }
            if (host.options.typeCheck && utils_1.isTypescript(load.address)) {
                return resolver.resolve(load.address)
                    .then(function (deps) {
                    var diags = typeChecker.check();
                    format_errors_1.formatErrors(diags, logger);
                    load.metadata.deps = deps.list
                        .filter(function (d) { return utils_1.isTypescript(d); })
                        .map(function (d) { return utils_1.isTypescriptDeclaration(d) ? d + "!" + __moduleName : d; });
                    if ((host.options.module === ts.ModuleKind.ES6) && !utils_1.isTypescriptDeclaration(load.address)) {
                        var importSource = deps.list
                            .filter(function (d) { return utils_1.isTypescript(d); })
                            .map(function (d) { return utils_1.isTypescriptDeclaration(d) ? d + "!" + __moduleName : d; })
                            .map(function (d) { return 'import "' + d + '"'; })
                            .join(';');
                        load.source = load.source + '\n' + importSource;
                    }
                    return load.source;
                });
            }
            else {
                return load.source;
            }
        });
    }
    exports_1("translate", translate);
    function bundle() {
        if (!factory)
            return [];
        return factory.then(function (_a) {
            var typeChecker = _a.typeChecker, host = _a.host;
            if (host.options.typeCheck) {
                var errors = typeChecker.forceCheck();
                format_errors_1.formatErrors(errors, logger);
                if ((host.options.typeCheck === "strict") && typeChecker.hasErrors())
                    throw new Error("Typescript compilation failed");
            }
            return [];
        });
    }
    exports_1("bundle", bundle);
    function validateOptions(options) {
        if ((options.module != ts.ModuleKind.System) && (options.module != ts.ModuleKind.ES6)) {
            logger.warn("transpiling to " + ts.ModuleKind[options.module] + ", consider setting module: \"system\" in typescriptOptions to transpile directly to System.register format");
        }
    }
    function _resolve(dep, parent) {
        if (!parent)
            parent = __moduleName;
        return System.normalize(dep, parent)
            .then(function (normalized) {
            normalized = utils_1.stripDoubleExtension(normalized);
            logger.debug("resolved " + normalized + " (" + parent + " -> " + dep + ")");
            return ts.normalizePath(normalized);
        });
    }
    function _fetch(address) {
        return System.fetch({ name: address, address: address, metadata: {} })
            .then(function (text) {
            logger.debug("fetched " + address);
            return text;
        });
    }
    function _lookup(address) {
        var metadata = {};
        return System.locate({ name: address, address: address, metadata: metadata })
            .then(function () {
            logger.debug("located " + address);
            return metadata;
        });
    }
    return {
        setters:[
            function (ts_1) {
                ts = ts_1;
            },
            function (logger_1_1) {
                logger_1 = logger_1_1;
            },
            function (factory_1_1) {
                factory_1 = factory_1_1;
            },
            function (format_errors_1_1) {
                format_errors_1 = format_errors_1_1;
            },
            function (utils_1_1) {
                utils_1 = utils_1_1;
            }],
        execute: function() {
            logger = new logger_1.default({ debug: false });
            factory = undefined;
        }
    }
});
