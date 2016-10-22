import { Scheduler } from '../Scheduler';
import { Observable } from '../Observable';
/**
 * @param delay
 * @param scheduler
 * @return {Observable<R>|WebSocketSubject<T>|Observable<T>}
 * @method auditTime
 * @owner Observable
 */
export declare function auditTime<T>(delay: number, scheduler?: Scheduler): Observable<T>;
export interface AuditTimeSignature<T> {
    (delay: number, scheduler?: Scheduler): Observable<T>;
}
