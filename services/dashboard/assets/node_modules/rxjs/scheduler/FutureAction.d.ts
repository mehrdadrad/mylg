import { Action } from './Action';
import { Scheduler } from '../Scheduler';
import { Subscription } from '../Subscription';
/**
 * We need this JSDoc comment for affecting ESDoc.
 * @ignore
 * @extends {Ignored}
 */
export declare class FutureAction<T> extends Subscription implements Action<T> {
    scheduler: Scheduler;
    work: (x?: T) => Subscription | void;
    id: number;
    state: T;
    delay: number;
    error: any;
    private pending;
    constructor(scheduler: Scheduler, work: (x?: T) => Subscription | void);
    execute(): void;
    schedule(state?: T, delay?: number): Action<T>;
    protected _schedule(state?: T, delay?: number): Action<T>;
    protected _unsubscribe(): void;
}
