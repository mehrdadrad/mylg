import { Operator } from './Operator';
import { Observer } from './Observer';
import { Observable } from './Observable';
import { Subscriber } from './Subscriber';
import { Subscription, ISubscription, TeardownLogic } from './Subscription';
/**
 * @class Subject<T>
 */
export declare class Subject<T> extends Observable<T> implements Observer<T>, ISubscription {
    protected destination: Observer<T>;
    protected source: Observable<T>;
    static create: Function;
    constructor(destination?: Observer<T>, source?: Observable<T>);
    observers: Observer<T>[];
    isUnsubscribed: boolean;
    protected isStopped: boolean;
    protected hasErrored: boolean;
    protected errorValue: any;
    protected dispatching: boolean;
    protected hasCompleted: boolean;
    lift<T, R>(operator: Operator<T, R>): Observable<T>;
    add(subscription: TeardownLogic): Subscription;
    remove(subscription: Subscription): void;
    unsubscribe(): void;
    protected _subscribe(subscriber: Subscriber<T>): TeardownLogic;
    protected _unsubscribe(): void;
    next(value: T): void;
    error(err?: any): void;
    complete(): void;
    asObservable(): Observable<T>;
    protected _next(value: T): void;
    protected _finalNext(value: T): void;
    protected _error(err: any): void;
    protected _finalError(err: any): void;
    protected _complete(): void;
    protected _finalComplete(): void;
    private throwIfUnsubscribed();
}
