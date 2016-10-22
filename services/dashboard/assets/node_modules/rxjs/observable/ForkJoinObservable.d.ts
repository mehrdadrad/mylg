import { Observable, SubscribableOrPromise } from '../Observable';
import { Subscriber } from '../Subscriber';
import { Subscription } from '../Subscription';
/**
 * We need this JSDoc comment for affecting ESDoc.
 * @extends {Ignored}
 * @hide true
 */
export declare class ForkJoinObservable<T> extends Observable<T> {
    private sources;
    private resultSelector;
    constructor(sources: Array<SubscribableOrPromise<any>>, resultSelector?: (...values: Array<any>) => T);
    /**
     * @param sources
     * @return {any}
     * @static true
     * @name forkJoin
     * @owner Observable
     */
    static create<T>(...sources: Array<SubscribableOrPromise<any> | Array<SubscribableOrPromise<any>> | ((...values: Array<any>) => any)>): Observable<T>;
    protected _subscribe(subscriber: Subscriber<any>): Subscription;
}
