import { Observable, SubscribableOrPromise } from '../Observable';
/**
 * @param durationSelector
 * @return {Observable<R>|WebSocketSubject<T>|Observable<T>}
 * @method throttle
 * @owner Observable
 */
export declare function throttle<T>(durationSelector: (value: T) => SubscribableOrPromise<number>): Observable<T>;
export interface ThrottleSignature<T> {
    (durationSelector: (value: T) => SubscribableOrPromise<number>): Observable<T>;
}
