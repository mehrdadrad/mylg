import { Scheduler } from '../Scheduler';
import { Observable } from '../Observable';
import { Subscriber } from '../Subscriber';
import { TeardownLogic } from '../Subscription';
/**
 * We need this JSDoc comment for affecting ESDoc.
 * @extends {Ignored}
 * @hide true
 */
export declare class PromiseObservable<T> extends Observable<T> {
    private promise;
    scheduler: Scheduler;
    value: T;
    /**
     * @param promise
     * @param scheduler
     * @return {PromiseObservable}
     * @static true
     * @name fromPromise
     * @owner Observable
     */
    static create<T>(promise: Promise<T>, scheduler?: Scheduler): Observable<T>;
    constructor(promise: Promise<T>, scheduler?: Scheduler);
    protected _subscribe(subscriber: Subscriber<T>): TeardownLogic;
}
