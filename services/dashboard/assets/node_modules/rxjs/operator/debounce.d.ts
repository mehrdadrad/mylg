import { Observable, SubscribableOrPromise } from '../Observable';
/**
 * Returns the source Observable delayed by the computed debounce duration,
 * with the duration lengthened if a new source item arrives before the delay
 * duration ends.
 * In practice, for each item emitted on the source, this operator holds the
 * latest item, waits for a silence as long as the `durationSelector` specifies,
 * and only then emits the latest source item on the result Observable.
 * @param {function} durationSelector function for computing the timeout duration for each item.
 * @return {Observable} an Observable the same as source Observable, but drops items.
 * @method debounce
 * @owner Observable
 */
export declare function debounce<T>(durationSelector: (value: T) => SubscribableOrPromise<number>): Observable<T>;
export interface DebounceSignature<T> {
    (durationSelector: (value: T) => SubscribableOrPromise<number>): Observable<T>;
}
