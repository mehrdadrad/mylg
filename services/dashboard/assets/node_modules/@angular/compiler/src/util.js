/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
import { StringWrapper, isArray, isBlank, isPresent, isPrimitive, isStrictStringMap } from './facade/lang';
import * as o from './output/output_ast';
export var MODULE_SUFFIX = '';
var CAMEL_CASE_REGEXP = /([A-Z])/g;
export function camelCaseToDashCase(input) {
    return StringWrapper.replaceAllMapped(input, CAMEL_CASE_REGEXP, function (m) { return '-' + m[1].toLowerCase(); });
}
export function splitAtColon(input, defaultValues) {
    return _splitAt(input, ':', defaultValues);
}
export function splitAtPeriod(input, defaultValues) {
    return _splitAt(input, '.', defaultValues);
}
function _splitAt(input, character, defaultValues) {
    var characterIndex = input.indexOf(character);
    if (characterIndex == -1)
        return defaultValues;
    return [input.slice(0, characterIndex).trim(), input.slice(characterIndex + 1).trim()];
}
export function sanitizeIdentifier(name) {
    return StringWrapper.replaceAll(name, /\W/g, '_');
}
export function visitValue(value, visitor, context) {
    if (isArray(value)) {
        return visitor.visitArray(value, context);
    }
    else if (isStrictStringMap(value)) {
        return visitor.visitStringMap(value, context);
    }
    else if (isBlank(value) || isPrimitive(value)) {
        return visitor.visitPrimitive(value, context);
    }
    else {
        return visitor.visitOther(value, context);
    }
}
export var ValueTransformer = (function () {
    function ValueTransformer() {
    }
    ValueTransformer.prototype.visitArray = function (arr, context) {
        var _this = this;
        return arr.map(function (value) { return visitValue(value, _this, context); });
    };
    ValueTransformer.prototype.visitStringMap = function (map, context) {
        var _this = this;
        var result = {};
        Object.keys(map).forEach(function (key) { result[key] = visitValue(map[key], _this, context); });
        return result;
    };
    ValueTransformer.prototype.visitPrimitive = function (value, context) { return value; };
    ValueTransformer.prototype.visitOther = function (value, context) { return value; };
    return ValueTransformer;
}());
export function assetUrl(pkg, path, type) {
    if (path === void 0) { path = null; }
    if (type === void 0) { type = 'src'; }
    if (path == null) {
        return "asset:@angular/lib/" + pkg + "/index";
    }
    else {
        return "asset:@angular/lib/" + pkg + "/src/" + path;
    }
}
export function createDiTokenExpression(token) {
    if (isPresent(token.value)) {
        return o.literal(token.value);
    }
    else if (token.identifierIsInstance) {
        return o.importExpr(token.identifier)
            .instantiate([], o.importType(token.identifier, [], [o.TypeModifier.Const]));
    }
    else {
        return o.importExpr(token.identifier);
    }
}
export var SyncAsyncResult = (function () {
    function SyncAsyncResult(syncResult, asyncResult) {
        if (asyncResult === void 0) { asyncResult = null; }
        this.syncResult = syncResult;
        this.asyncResult = asyncResult;
        if (!asyncResult) {
            this.asyncResult = Promise.resolve(syncResult);
        }
    }
    return SyncAsyncResult;
}());
//# sourceMappingURL=util.js.map