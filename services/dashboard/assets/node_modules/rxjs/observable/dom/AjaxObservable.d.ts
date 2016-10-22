import { Observable } from '../../Observable';
import { Subscriber } from '../../Subscriber';
import { TeardownLogic } from '../../Subscription';
export interface AjaxRequest {
    url?: string;
    body?: any;
    user?: string;
    async?: boolean;
    method: string;
    headers?: Object;
    timeout?: number;
    password?: string;
    hasContent?: boolean;
    crossDomain?: boolean;
    createXHR?: () => XMLHttpRequest;
    progressSubscriber?: Subscriber<any>;
    resultSelector?: <T>(response: AjaxResponse) => T;
    responseType?: string;
}
export interface AjaxCreationMethod {
    <T>(urlOrRequest: string | AjaxRequest): Observable<T>;
    get<T>(url: string, resultSelector?: (response: AjaxResponse) => T, headers?: Object): Observable<T>;
    post<T>(url: string, body?: any, headers?: Object): Observable<T>;
    put<T>(url: string, body?: any, headers?: Object): Observable<T>;
    delete<T>(url: string, headers?: Object): Observable<T>;
    getJSON<T, R>(url: string, resultSelector?: (data: T) => R, headers?: Object): Observable<R>;
}
export declare function ajaxGet<T>(url: string, resultSelector?: (response: AjaxResponse) => T, headers?: Object): AjaxObservable<T>;
export declare function ajaxPost<T>(url: string, body?: any, headers?: Object): Observable<T>;
export declare function ajaxDelete<T>(url: string, headers?: Object): Observable<T>;
export declare function ajaxPut<T>(url: string, body?: any, headers?: Object): Observable<T>;
export declare function ajaxGetJSON<T, R>(url: string, resultSelector?: (data: T) => R, headers?: Object): Observable<R>;
/**
 * We need this JSDoc comment for affecting ESDoc.
 * @extends {Ignored}
 * @hide true
 */
export declare class AjaxObservable<T> extends Observable<T> {
    /**
     * Creates an observable for an Ajax request with either a request object with
     * url, headers, etc or a string for a URL.
     *
     * @example
     * source = Rx.Observable.ajax('/products');
     * source = Rx.Observable.ajax( url: 'products', method: 'GET' });
     *
     * @param {string|Object} request Can be one of the following:
     *   A string of the URL to make the Ajax call.
     *   An object with the following properties
     *   - url: URL of the request
     *   - body: The body of the request
     *   - method: Method of the request, such as GET, POST, PUT, PATCH, DELETE
     *   - async: Whether the request is async
     *   - headers: Optional headers
     *   - crossDomain: true if a cross domain request, else false
     *   - createXHR: a function to override if you need to use an alternate
     *   XMLHttpRequest implementation.
     *   - resultSelector: a function to use to alter the output value type of
     *   the Observable. Gets {@link AjaxResponse} as an argument.
     * @return {Observable} An observable sequence containing the XMLHttpRequest.
     * @static true
     * @name ajax
     * @owner Observable
    */
    static _create_stub(): void;
    static create: AjaxCreationMethod;
    private request;
    constructor(urlOrRequest: string | AjaxRequest);
    protected _subscribe(subscriber: Subscriber<T>): TeardownLogic;
}
/**
 * We need this JSDoc comment for affecting ESDoc.
 * @ignore
 * @extends {Ignored}
 */
export declare class AjaxSubscriber<T> extends Subscriber<Event> {
    request: AjaxRequest;
    private xhr;
    private resultSelector;
    private done;
    constructor(destination: Subscriber<T>, request: AjaxRequest);
    next(e: Event): void;
    private send();
    private serializeBody(body, contentType);
    private setHeaders(xhr, headers);
    private setupEvents(xhr, request);
    unsubscribe(): void;
}
/**
 * A normalized AJAX response.
 *
 * @see {@link ajax}
 *
 * @class AjaxResponse
 */
export declare class AjaxResponse {
    originalEvent: Event;
    xhr: XMLHttpRequest;
    request: AjaxRequest;
    /** @type {number} The HTTP status code */
    status: number;
    /** @type {string|ArrayBuffer|Document|object|any} The response data */
    response: any;
    /** @type {string} The raw responseText */
    responseText: string;
    /** @type {string} The responseType (e.g. 'json', 'arraybuffer', or 'xml') */
    responseType: string;
    constructor(originalEvent: Event, xhr: XMLHttpRequest, request: AjaxRequest);
}
/**
 * A normalized AJAX error.
 *
 * @see {@link ajax}
 *
 * @class AjaxError
 */
export declare class AjaxError extends Error {
    /** @type {XMLHttpRequest} The XHR instance associated with the error */
    xhr: XMLHttpRequest;
    /** @type {AjaxRequest} The AjaxRequest associated with the error */
    request: AjaxRequest;
    /** @type {number} The HTTP status code */
    status: number;
    constructor(message: string, xhr: XMLHttpRequest, request: AjaxRequest);
}
/**
 * @see {@link ajax}
 *
 * @class AjaxTimeoutError
 */
export declare class AjaxTimeoutError extends AjaxError {
    constructor(xhr: XMLHttpRequest, request: AjaxRequest);
}
