import { Observable } from '../Observable';
/**
 * @param notifier
 * @return {Observable<R>|WebSocketSubject<T>|Observable<T>}
 * @method takeUntil
 * @owner Observable
 */
export declare function takeUntil<T>(notifier: Observable<any>): Observable<T>;
export interface TakeUntilSignature<T> {
    (notifier: Observable<any>): Observable<T>;
}
