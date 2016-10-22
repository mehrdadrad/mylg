import { Scheduler } from '../Scheduler';
import { QueueAction } from './QueueAction';
import { Subscription } from '../Subscription';
import { Action } from './Action';
export declare class QueueScheduler implements Scheduler {
    active: boolean;
    actions: QueueAction<any>[];
    scheduledId: number;
    now(): number;
    flush(): void;
    schedule<T>(work: (x?: T) => Subscription | void, delay?: number, state?: T): Subscription;
    scheduleNow<T>(work: (x?: T) => Subscription | void, state?: T): Action<T>;
    scheduleLater<T>(work: (x?: T) => Subscription | void, delay: number, state?: T): Action<T>;
}
