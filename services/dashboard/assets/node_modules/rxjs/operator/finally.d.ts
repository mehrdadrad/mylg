import { Observable } from '../Observable';
/**
 * Returns an Observable that mirrors the source Observable, but will call a specified function when
 * the source terminates on complete or error.
 * @param {function} finallySelector function to be called when source terminates.
 * @return {Observable} an Observable that mirrors the source, but will call the specified function on termination.
 * @method finally
 * @owner Observable
 */
export declare function _finally<T>(finallySelector: () => void): Observable<T>;
export interface FinallySignature<T> {
    <T>(finallySelector: () => void): Observable<T>;
}
