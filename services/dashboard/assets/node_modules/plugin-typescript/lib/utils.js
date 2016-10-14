System.register(["typescript"], function(exports_1, context_1) {
    "use strict";
    var __moduleName = context_1 && context_1.id;
    var ts;
    var typescriptRegex, javascriptRegex, mapRegex, declarationRegex, htmlRegex;
    function isAbsolute(filename) {
        return (filename[0] === '/');
    }
    exports_1("isAbsolute", isAbsolute);
    function isRelative(filename) {
        return (filename[0] === '.');
    }
    exports_1("isRelative", isRelative);
    function isAmbientImport(filename) {
        return (isAmbient(filename) && !isTypescriptDeclaration(filename));
    }
    exports_1("isAmbientImport", isAmbientImport);
    function isAmbientReference(filename) {
        return (isAmbient(filename) && isTypescriptDeclaration(filename));
    }
    exports_1("isAmbientReference", isAmbientReference);
    function isAmbient(filename) {
        return (!isRelative(filename) && !isAbsolute(filename));
    }
    exports_1("isAmbient", isAmbient);
    function isTypescript(filename) {
        return typescriptRegex.test(filename);
    }
    exports_1("isTypescript", isTypescript);
    function isJavaScript(filename) {
        return javascriptRegex.test(filename);
    }
    exports_1("isJavaScript", isJavaScript);
    function isSourceMap(filename) {
        return mapRegex.test(filename);
    }
    exports_1("isSourceMap", isSourceMap);
    function isTypescriptDeclaration(filename) {
        return declarationRegex.test(filename);
    }
    exports_1("isTypescriptDeclaration", isTypescriptDeclaration);
    function isHtml(filename) {
        return htmlRegex.test(filename);
    }
    exports_1("isHtml", isHtml);
    function tsToJs(tsFile) {
        return tsFile.replace(typescriptRegex, '.js');
    }
    exports_1("tsToJs", tsToJs);
    function tsToJsMap(tsFile) {
        return tsFile.replace(typescriptRegex, '.js.map');
    }
    exports_1("tsToJsMap", tsToJsMap);
    function jsToDts(jsFile) {
        return jsFile.replace(javascriptRegex, '.d.ts');
    }
    exports_1("jsToDts", jsToDts);
    function stripDoubleExtension(normalized) {
        var parts = normalized.split('.');
        if (parts.length > 1) {
            var extensions = ["js", "jsx", "ts", "tsx", "json"];
            if (extensions.indexOf(parts[parts.length - 2]) >= 0) {
                return parts.slice(0, -1).join('.');
            }
        }
        return normalized;
    }
    exports_1("stripDoubleExtension", stripDoubleExtension);
    function hasError(diags) {
        return diags.some(function (diag) { return (diag.category === ts.DiagnosticCategory.Error); });
    }
    exports_1("hasError", hasError);
    return {
        setters:[
            function (ts_1) {
                ts = ts_1;
            }],
        execute: function() {
            typescriptRegex = /\.tsx?$/i;
            javascriptRegex = /\.js$/i;
            mapRegex = /\.map$/i;
            declarationRegex = /\.d\.tsx?$/i;
            htmlRegex = /\.html$/i;
        }
    }
});
