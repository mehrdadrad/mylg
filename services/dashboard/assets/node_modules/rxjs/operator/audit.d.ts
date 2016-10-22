import { Observable, SubscribableOrPromise } from '../Observable';
/**
 * @param durationSelector
 * @return {Observable<R>|WebSocketSubject<T>|Observable<T>}
 * @method audit
 * @owner Observable
 */
export declare function audit<T>(durationSelector: (value: T) => SubscribableOrPromise<any>): Observable<T>;
export interface AuditSignature<T> {
    (durationSelector: (value: T) => SubscribableOrPromise<any>): Observable<T>;
}
