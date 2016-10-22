/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
import { isString } from '../src/facade/lang';
import { RequestMethod } from './enums';
export function normalizeMethodName(method) {
    if (isString(method)) {
        var originalMethod = method;
        method = method
            .replace(/(\w)(\w*)/g, function (g0, g1, g2) { return g1.toUpperCase() + g2.toLowerCase(); });
        method = RequestMethod[method];
        if (typeof method !== 'number')
            throw new Error("Invalid request method. The method \"" + originalMethod + "\" is not supported.");
    }
    return method;
}
export var isSuccess = function (status) { return (status >= 200 && status < 300); };
export function getResponseURL(xhr) {
    if ('responseURL' in xhr) {
        return xhr.responseURL;
    }
    if (/^X-Request-URL:/m.test(xhr.getAllResponseHeaders())) {
        return xhr.getResponseHeader('X-Request-URL');
    }
    return;
}
export function stringToArrayBuffer(input) {
    var view = new Uint16Array(input.length);
    for (var i = 0, strLen = input.length; i < strLen; i++) {
        view[i] = input.charCodeAt(i);
    }
    return view.buffer;
}
export { isJsObject } from '../src/facade/lang';
//# sourceMappingURL=http_utils.js.map