import { Observable } from '../Observable';
import { Scheduler } from '../Scheduler';
/**
 * @param delay
 * @param scheduler
 * @return {Observable<R>|WebSocketSubject<T>|Observable<T>}
 * @method sampleTime
 * @owner Observable
 */
export declare function sampleTime<T>(delay: number, scheduler?: Scheduler): Observable<T>;
export interface SampleTimeSignature<T> {
    (delay: number, scheduler?: Scheduler): Observable<T>;
}
