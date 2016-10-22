import { Observable } from '../Observable';
/**
 * @throws {ArgumentOutOfRangeError} When using `take(i)`, it delivers an
 * ArgumentOutOrRangeError to the Observer's `error` callback if `i < 0`.
 * @param total
 * @return {any}
 * @method take
 * @owner Observable
 */
export declare function take<T>(total: number): Observable<T>;
export interface TakeSignature<T> {
    (total: number): Observable<T>;
}
