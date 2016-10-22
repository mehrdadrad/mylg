import { Observable } from '../Observable';
/**
 * Returns an Observable that transforms Notification objects into the items or notifications they represent.
 *
 * @see {@link Notification}
 *
 * @return {Observable} an Observable that emits items and notifications embedded in Notification objects emitted by the source Observable.
 * @method dematerialize
 * @owner Observable
 */
export declare function dematerialize<T>(): Observable<any>;
export interface DematerializeSignature<T> {
    <R>(): Observable<R>;
}
