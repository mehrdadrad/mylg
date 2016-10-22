import { Scheduler } from '../Scheduler';
import { Observable } from '../Observable';
import { TeardownLogic } from '../Subscription';
import { Subscriber } from '../Subscriber';
/**
 * We need this JSDoc comment for affecting ESDoc.
 * @extends {Ignored}
 * @hide true
 */
export declare class IteratorObservable<T> extends Observable<T> {
    private iterator;
    static create<T>(iterator: any, project?: ((x?: any, i?: number) => T) | any, thisArg?: any | Scheduler, scheduler?: Scheduler): IteratorObservable<{}>;
    static dispatch(state: any): void;
    private thisArg;
    private project;
    private scheduler;
    constructor(iterator: any, project?: ((x?: any, i?: number) => T) | any, thisArg?: any | Scheduler, scheduler?: Scheduler);
    protected _subscribe(subscriber: Subscriber<T>): TeardownLogic;
}
