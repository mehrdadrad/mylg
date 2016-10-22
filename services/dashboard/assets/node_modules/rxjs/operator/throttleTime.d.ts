import { Scheduler } from '../Scheduler';
import { Observable } from '../Observable';
/**
 * @param delay
 * @param scheduler
 * @return {Observable<R>|WebSocketSubject<T>|Observable<T>}
 * @method throttleTime
 * @owner Observable
 */
export declare function throttleTime<T>(delay: number, scheduler?: Scheduler): Observable<T>;
export interface ThrottleTimeSignature<T> {
    (dueTime: number, scheduler?: Scheduler): Observable<T>;
}
