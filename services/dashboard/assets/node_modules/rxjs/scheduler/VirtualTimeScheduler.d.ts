import { Scheduler } from '../Scheduler';
import { Subscription } from '../Subscription';
import { Action } from './Action';
export declare class VirtualTimeScheduler implements Scheduler {
    actions: Action<any>[];
    active: boolean;
    scheduledId: number;
    index: number;
    sorted: boolean;
    frame: number;
    maxFrames: number;
    protected static frameTimeFactor: number;
    now(): number;
    flush(): void;
    addAction<T>(action: Action<T>): void;
    schedule<T>(work: (x?: T) => Subscription | void, delay?: number, state?: T): Subscription;
}
