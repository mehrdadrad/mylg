import { Observable } from '../Observable';
/**
 * @param predicate
 * @return {Observable<R>|WebSocketSubject<T>|Observable<T>}
 * @method takeWhile
 * @owner Observable
 */
export declare function takeWhile<T>(predicate: (value: T, index: number) => boolean): Observable<T>;
export interface TakeWhileSignature<T> {
    (predicate: (value: T, index: number) => boolean): Observable<T>;
}
