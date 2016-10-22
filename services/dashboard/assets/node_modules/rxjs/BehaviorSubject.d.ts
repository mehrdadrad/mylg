import { Subject } from './Subject';
import { Subscriber } from './Subscriber';
import { TeardownLogic } from './Subscription';
/**
 * @class BehaviorSubject<T>
 */
export declare class BehaviorSubject<T> extends Subject<T> {
    private _value;
    constructor(_value: T);
    getValue(): T;
    value: T;
    protected _subscribe(subscriber: Subscriber<T>): TeardownLogic;
    protected _next(value: T): void;
    protected _error(err: any): void;
}
