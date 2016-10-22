import { Observable } from '../Observable';
/**
 * @throws {ArgumentOutOfRangeError} When using `takeLast(i)`, it delivers an
 * ArgumentOutOrRangeError to the Observer's `error` callback if `i < 0`.
 * @param total
 * @return {any}
 * @method takeLast
 * @owner Observable
 */
export declare function takeLast<T>(total: number): Observable<T>;
export interface TakeLastSignature<T> {
    (total: number): Observable<T>;
}
