import { Scheduler } from '../Scheduler';
import { Observable } from '../Observable';
/**
 * @param due
 * @param withObservable
 * @param scheduler
 * @return {Observable<R>|WebSocketSubject<T>|Observable<T>}
 * @method timeoutWith
 * @owner Observable
 */
export declare function timeoutWith<T, R>(due: number | Date, withObservable: Observable<R>, scheduler?: Scheduler): Observable<T | R>;
export interface TimeoutWithSignature<T> {
    (due: number | Date, withObservable: Observable<T>, scheduler?: Scheduler): Observable<T>;
    <R>(due: number | Date, withObservable: Observable<R>, scheduler?: Scheduler): Observable<T | R>;
}
