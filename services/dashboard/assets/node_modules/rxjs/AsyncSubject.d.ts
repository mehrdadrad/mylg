import { Subject } from './Subject';
import { Subscriber } from './Subscriber';
import { TeardownLogic } from './Subscription';
/**
 * @class AsyncSubject<T>
 */
export declare class AsyncSubject<T> extends Subject<T> {
    value: T;
    hasNext: boolean;
    protected _subscribe(subscriber: Subscriber<any>): TeardownLogic;
    protected _next(value: T): void;
    protected _complete(): void;
}
