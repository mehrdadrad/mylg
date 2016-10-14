"use strict";
var core_1 = require('@angular/core');
var http_1 = require('@angular/http');
var Observable_1 = require('rxjs/Observable');
require('rxjs/add/operator/delay');
var http_status_codes_1 = require('./http-status-codes');
////////////  HELPERS ///////////
/**
 * Create an error Response from an HTTP status code and error message
 */
function createErrorResponse(status, message) {
    return new http_1.ResponseOptions({
        body: { 'error': "" + message },
        headers: new http_1.Headers({ 'Content-Type': 'application/json' }),
        status: status
    });
}
exports.createErrorResponse = createErrorResponse;
/**
 * Create an Observable response from response options:
 */
function createObservableResponse(resOptions) {
    resOptions = setStatusText(resOptions);
    var res = new http_1.Response(resOptions);
    return new Observable_1.Observable(function (responseObserver) {
        if (isSuccess(res.status)) {
            responseObserver.next(res);
            responseObserver.complete();
        }
        else {
            responseObserver.error(res);
        }
        return function () { }; // unsubscribe function
    });
}
exports.createObservableResponse = createObservableResponse;
/**
* Interface for a class that creates an in-memory database
*
* Its `createDb` method creates a hash of named collections that represents the database
*
* For maximum flexibility, the service may define HTTP method overrides.
* Such methods must match the spelling of an HTTP method in lower case (e.g, "get").
* If a request has a matching method, it will be called as in
* `get(info: requestInfo, db: {})` where `db` is the database object described above.
*/
var InMemoryDbService = (function () {
    function InMemoryDbService() {
    }
    return InMemoryDbService;
}());
exports.InMemoryDbService = InMemoryDbService;
/**
*  InMemoryBackendService configuration options
*  Usage:
*    InMemoryWebApiModule.forRoot(InMemHeroService, {delay: 600})
*
*  or if providing separately:
*    provide(InMemoryBackendConfig, {useValue: {delay: 600}}),
*/
var InMemoryBackendConfig = (function () {
    function InMemoryBackendConfig(config) {
        if (config === void 0) { config = {}; }
        Object.assign(this, {
            // default config:
            caseSensitiveSearch: false,
            defaultResponseOptions: new http_1.BaseResponseOptions(),
            delay: 500,
            delete404: false,
            passThruUnknownUrl: false,
            host: '',
            rootPath: ''
        }, config);
    }
    return InMemoryBackendConfig;
}());
exports.InMemoryBackendConfig = InMemoryBackendConfig;
/**
 * Returns true if the the Http Status Code is 200-299 (success)
 */
function isSuccess(status) { return status >= 200 && status < 300; }
exports.isSuccess = isSuccess;
;
/**
 * Set the status text in a response:
 */
function setStatusText(options) {
    try {
        var statusCode = http_status_codes_1.STATUS_CODE_INFO[options.status];
        options['statusText'] = statusCode ? statusCode.text : 'Unknown Status';
        return options;
    }
    catch (err) {
        return new http_1.ResponseOptions({
            status: http_status_codes_1.STATUS.INTERNAL_SERVER_ERROR,
            statusText: 'Invalid Server Operation'
        });
    }
}
exports.setStatusText = setStatusText;
////////////  InMemoryBackendService ///////////
/**
 * Simulate the behavior of a RESTy web api
 * backed by the simple in-memory data store provided by the injected InMemoryDataService service.
 * Conforms mostly to behavior described here:
 * http://www.restapitutorial.com/lessons/httpmethods.html
 *
 * ### Usage
 *
 * Create `InMemoryDataService` class that implements `InMemoryDataService`.
 * Call `forRoot` static method with this service class and optional configuration object:
 * ```
 * // other imports
 * import { HttpModule }           from '@angular/http';
 * import { InMemoryWebApiModule } from 'angular-in-memory-web-api';
 *
 * import { InMemHeroService, inMemConfig } from '../api/in-memory-hero.service';
 * @NgModule({
 *  imports: [
 *    HttpModule,
 *    InMemoryWebApiModule.forRoot(InMemHeroService, inMemConfig),
 *    ...
 *  ],
 *  ...
 * })
 * export class AppModule { ... }
 * ```
 */
var InMemoryBackendService = (function () {
    function InMemoryBackendService(injector, inMemDbService, config) {
        this.injector = injector;
        this.inMemDbService = inMemDbService;
        this.config = new InMemoryBackendConfig();
        this.resetDb();
        var loc = this.getLocation('./');
        this.config.host = loc.host;
        this.config.rootPath = loc.pathname;
        Object.assign(this.config, config || {});
        this.setPassThruBackend();
    }
    InMemoryBackendService.prototype.createConnection = function (req) {
        var response = this.handleRequest(req);
        return {
            readyState: http_1.ReadyState.Done,
            request: req,
            response: response
        };
    };
    ////  protected /////
    /**
     * Process Request and return an Observable of Http Response object
     * in the manner of a RESTy web api.
     *
     * Expect URI pattern in the form :base/:collectionName/:id?
     * Examples:
     *   // for store with a 'characters' collection
     *   GET api/characters          // all characters
     *   GET api/characters/42       // the character with id=42
     *   GET api/characters?name=^j  // 'j' is a regex; returns characters whose name starts with 'j' or 'J'
     *   GET api/characters.json/42  // ignores the ".json"
     *
     * Also accepts
     *   "commands":
     *     POST "resetDb",
     *     GET/POST "config"" - get or (re)set the config
     *
     *   HTTP overrides:
     *     If the injected inMemDbService defines an HTTP method (lowercase)
     *     The request is forwarded to that method as in
     *     `inMemDbService.get(httpMethodInterceptorArgs)`
     *     which must return an `Observable<Response>`
     */
    InMemoryBackendService.prototype.handleRequest = function (req) {
        var parsed = this.inMemDbService['parseUrl'] ?
            // parse with override method
            this.inMemDbService['parseUrl'](req.url) :
            // parse with default url parser
            this.parseUrl(req.url);
        var base = parsed.base, collectionName = parsed.collectionName, id = parsed.id, query = parsed.query, resourceUrl = parsed.resourceUrl;
        var collection = this.db[collectionName];
        var reqInfo = {
            req: req,
            base: base,
            collection: collection,
            collectionName: collectionName,
            headers: new http_1.Headers({ 'Content-Type': 'application/json' }),
            id: this.parseId(collection, id),
            query: query,
            resourceUrl: resourceUrl
        };
        var reqMethodName = http_1.RequestMethod[req.method || 0].toLowerCase();
        var resOptions;
        try {
            if ('commands' === reqInfo.base.toLowerCase()) {
                return this.commands(reqInfo);
            }
            else if (this.inMemDbService[reqMethodName]) {
                // If service has an interceptor for an HTTP method, call it
                var interceptorArgs = {
                    requestInfo: reqInfo,
                    db: this.db,
                    config: this.config,
                    passThruBackend: this.passThruBackend
                };
                // The result which must be Observable<Response>
                return this.inMemDbService[reqMethodName](interceptorArgs);
            }
            else if (reqInfo.collection) {
                return this.collectionHandler(reqInfo);
            }
            else if (this.passThruBackend) {
                // Passes request thru to a "real" backend which returns an Observable<Response>
                // BAIL OUT with this Observable<Response>
                return this.passThruBackend.createConnection(req).response;
            }
            else {
                resOptions = createErrorResponse(http_status_codes_1.STATUS.NOT_FOUND, "Collection '" + collectionName + "' not found");
                return createObservableResponse(resOptions);
            }
        }
        catch (error) {
            var err = error.message || error;
            resOptions = createErrorResponse(http_status_codes_1.STATUS.INTERNAL_SERVER_ERROR, "" + err);
            return createObservableResponse(resOptions);
        }
    };
    /**
     * Apply query/search parameters as a filter over the collection
     * This impl only supports RegExp queries on string properties of the collection
     * ANDs the conditions together
     */
    InMemoryBackendService.prototype.applyQuery = function (collection, query) {
        // extract filtering conditions - {propertyName, RegExps) - from query/search parameters
        var conditions = [];
        var caseSensitive = this.config.caseSensitiveSearch ? undefined : 'i';
        query.paramsMap.forEach(function (value, name) {
            value.forEach(function (v) { return conditions.push({ name: name, rx: new RegExp(decodeURI(v), caseSensitive) }); });
        });
        var len = conditions.length;
        if (!len) {
            return collection;
        }
        // AND the RegExp conditions
        return collection.filter(function (row) {
            var ok = true;
            var i = len;
            while (ok && i) {
                i -= 1;
                var cond = conditions[i];
                ok = cond.rx.test(row[cond.name]);
            }
            return ok;
        });
    };
    InMemoryBackendService.prototype.clone = function (data) {
        return JSON.parse(JSON.stringify(data));
    };
    InMemoryBackendService.prototype.collectionHandler = function (reqInfo) {
        var req = reqInfo.req;
        var resOptions;
        switch (req.method) {
            case http_1.RequestMethod.Get:
                resOptions = this.get(reqInfo);
                break;
            case http_1.RequestMethod.Post:
                resOptions = this.post(reqInfo);
                break;
            case http_1.RequestMethod.Put:
                resOptions = this.put(reqInfo);
                break;
            case http_1.RequestMethod.Delete:
                resOptions = this.delete(reqInfo);
                break;
            default:
                resOptions = createErrorResponse(http_status_codes_1.STATUS.METHOD_NOT_ALLOWED, 'Method not allowed');
                break;
        }
        return createObservableResponse(resOptions);
    };
    /**
     * When the `base`="commands", the `collectionName` is the command
     * Example URLs:
     *   commands/resetdb   // Reset the "database" to its original state
     *   commands/config (GET) // Return this service's config object
     *   commands/config (!GET) // Update the config (e.g. delay)
     *
     * Usage:
     *   http.post('commands/resetdb', null);
     *   http.get('commands/config');
     *   http.post('commands/config', '{"delay":1000}');
     */
    InMemoryBackendService.prototype.commands = function (reqInfo) {
        var command = reqInfo.collectionName.toLowerCase();
        var method = reqInfo.req.method;
        var resOptions;
        switch (command) {
            case 'resetdb':
                this.resetDb();
                resOptions = new http_1.ResponseOptions({ status: http_status_codes_1.STATUS.OK });
                break;
            case 'config':
                if (method === http_1.RequestMethod.Get) {
                    resOptions = new http_1.ResponseOptions({
                        body: this.clone(this.config),
                        status: http_status_codes_1.STATUS.OK
                    });
                }
                else {
                    // Be nice ... any other method is a config update
                    var body = JSON.parse(reqInfo.req.text() || '{}');
                    Object.assign(this.config, body);
                    this.setPassThruBackend();
                    resOptions = new http_1.ResponseOptions({ status: http_status_codes_1.STATUS.NO_CONTENT });
                }
                break;
            default:
                resOptions = createErrorResponse(http_status_codes_1.STATUS.INTERNAL_SERVER_ERROR, "Unknown command \"" + command + "\"");
        }
        return createObservableResponse(resOptions);
    };
    InMemoryBackendService.prototype.delete = function (_a) {
        var id = _a.id, collection = _a.collection, collectionName = _a.collectionName, headers = _a.headers;
        if (!id) {
            return createErrorResponse(http_status_codes_1.STATUS.NOT_FOUND, "Missing \"" + collectionName + "\" id");
        }
        var exists = this.removeById(collection, id);
        return new http_1.ResponseOptions({
            headers: headers,
            status: (exists || !this.config.delete404) ? http_status_codes_1.STATUS.NO_CONTENT : http_status_codes_1.STATUS.NOT_FOUND
        });
    };
    InMemoryBackendService.prototype.findById = function (collection, id) {
        return collection.find(function (item) { return item.id === id; });
    };
    InMemoryBackendService.prototype.genId = function (collection) {
        // assumes numeric ids
        var maxId = 0;
        collection.reduce(function (prev, item) {
            maxId = Math.max(maxId, typeof item.id === 'number' ? item.id : maxId);
        }, null);
        return maxId + 1;
    };
    InMemoryBackendService.prototype.get = function (_a) {
        var id = _a.id, query = _a.query, collection = _a.collection, collectionName = _a.collectionName, headers = _a.headers;
        var data = collection;
        if (id) {
            data = this.findById(collection, id);
        }
        else if (query) {
            data = this.applyQuery(collection, query);
        }
        if (!data) {
            return createErrorResponse(http_status_codes_1.STATUS.NOT_FOUND, "'" + collectionName + "' with id='" + id + "' not found");
        }
        return new http_1.ResponseOptions({
            body: { data: this.clone(data) },
            headers: headers,
            status: http_status_codes_1.STATUS.OK
        });
    };
    InMemoryBackendService.prototype.getLocation = function (href) {
        var l = document.createElement('a');
        l.href = href;
        return l;
    };
    ;
    InMemoryBackendService.prototype.indexOf = function (collection, id) {
        return collection.findIndex(function (item) { return item.id === id; });
    };
    // tries to parse id as number if collection item.id is a number.
    // returns the original param id otherwise.
    InMemoryBackendService.prototype.parseId = function (collection, id) {
        if (!collection || !id) {
            return null;
        }
        var isNumberId = collection[0] && typeof collection[0].id === 'number';
        if (isNumberId) {
            var idNum = parseFloat(id);
            return isNaN(idNum) ? id : idNum;
        }
        return id;
    };
    InMemoryBackendService.prototype.parseUrl = function (url) {
        try {
            var loc = this.getLocation(url);
            var drop = this.config.rootPath.length;
            var urlRoot = '';
            if (loc.host !== this.config.host) {
                // url for a server on a different host!
                // assume it's collection is actually here too.
                drop = 1; // the leading slash
                urlRoot = loc.protocol + '//' + loc.host + '/';
            }
            var path = loc.pathname.substring(drop);
            var _a = path.split('/'), base = _a[0], collectionName = _a[1], id = _a[2];
            var resourceUrl = urlRoot + base + '/' + collectionName + '/';
            collectionName = collectionName.split('.')[0]; // ignore anything after the '.', e.g., '.json'
            var query = loc.search && new http_1.URLSearchParams(loc.search.substr(1));
            return { base: base, collectionName: collectionName, id: id, query: query, resourceUrl: resourceUrl };
        }
        catch (err) {
            var msg = "unable to parse url '" + url + "'; original error: " + err.message;
            throw new Error(msg);
        }
    };
    InMemoryBackendService.prototype.post = function (_a) {
        var collection = _a.collection, headers = _a.headers, id = _a.id, req = _a.req, resourceUrl = _a.resourceUrl;
        var item = JSON.parse(req.text());
        if (!item.id) {
            item.id = id || this.genId(collection);
        }
        // ignore the request id, if any. Alternatively,
        // could reject request if id differs from item.id
        id = item.id;
        var existingIx = this.indexOf(collection, id);
        if (existingIx > -1) {
            collection[existingIx] = item;
            return new http_1.ResponseOptions({
                headers: headers,
                status: http_status_codes_1.STATUS.NO_CONTENT
            });
        }
        else {
            collection.push(item);
            headers.set('Location', resourceUrl + '/' + id);
            return new http_1.ResponseOptions({
                headers: headers,
                body: { data: this.clone(item) },
                status: http_status_codes_1.STATUS.CREATED
            });
        }
    };
    InMemoryBackendService.prototype.put = function (_a) {
        var id = _a.id, collection = _a.collection, collectionName = _a.collectionName, headers = _a.headers, req = _a.req;
        var item = JSON.parse(req.text());
        if (!id) {
            return createErrorResponse(http_status_codes_1.STATUS.NOT_FOUND, "Missing '" + collectionName + "' id");
        }
        if (id !== item.id) {
            return createErrorResponse(http_status_codes_1.STATUS.BAD_REQUEST, "\"" + collectionName + "\" id does not match item.id");
        }
        var existingIx = this.indexOf(collection, id);
        if (existingIx > -1) {
            collection[existingIx] = item;
            return new http_1.ResponseOptions({
                headers: headers,
                status: http_status_codes_1.STATUS.NO_CONTENT // successful; no content
            });
        }
        else {
            collection.push(item);
            return new http_1.ResponseOptions({
                body: { data: this.clone(item) },
                headers: headers,
                status: http_status_codes_1.STATUS.CREATED
            });
        }
    };
    InMemoryBackendService.prototype.removeById = function (collection, id) {
        var ix = this.indexOf(collection, id);
        if (ix > -1) {
            collection.splice(ix, 1);
            return true;
        }
        return false;
    };
    /**
     * Reset the "database" to its original state
     */
    InMemoryBackendService.prototype.resetDb = function () {
        this.db = this.inMemDbService.createDb();
    };
    InMemoryBackendService.prototype.setPassThruBackend = function () {
        this.passThruBackend = undefined;
        if (this.config.passThruUnknownUrl) {
            try {
                // copied from @angular/http/backends/xhr_backend
                var browserXhr = this.injector.get(http_1.BrowserXhr);
                var baseResponseOptions = this.injector.get(http_1.ResponseOptions);
                var xsrfStrategy = this.injector.get(http_1.XSRFStrategy);
                this.passThruBackend = new http_1.XHRBackend(browserXhr, baseResponseOptions, xsrfStrategy);
            }
            catch (ex) {
                ex.message = 'Cannot create passThru404 backend; ' + (ex.message || '');
                throw ex;
            }
        }
    };
    /** @nocollapse */
    InMemoryBackendService.ctorParameters = [
        { type: core_1.Injector, },
        { type: InMemoryDbService, },
        { type: undefined, decorators: [{ type: core_1.Inject, args: [InMemoryBackendConfig,] }, { type: core_1.Optional },] },
    ];
    return InMemoryBackendService;
}());
exports.InMemoryBackendService = InMemoryBackendService;
//# sourceMappingURL=in-memory-backend.service.js.map