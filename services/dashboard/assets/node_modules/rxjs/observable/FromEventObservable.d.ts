import { Observable } from '../Observable';
import { Subscriber } from '../Subscriber';
export declare type NodeStyleEventEmmitter = {
    addListener: (eventName: string, handler: Function) => void;
    removeListener: (eventName: string, handler: Function) => void;
};
export declare type JQueryStyleEventEmitter = {
    on: (eventName: string, handler: Function) => void;
    off: (eventName: string, handler: Function) => void;
};
export declare type EventTargetLike = EventTarget | NodeStyleEventEmmitter | JQueryStyleEventEmitter | NodeList | HTMLCollection;
/**
 * We need this JSDoc comment for affecting ESDoc.
 * @extends {Ignored}
 * @hide true
 */
export declare class FromEventObservable<T, R> extends Observable<T> {
    private sourceObj;
    private eventName;
    private selector;
    /**
     * @param sourceObj
     * @param eventName
     * @param selector
     * @return {FromEventObservable}
     * @static true
     * @name fromEvent
     * @owner Observable
     */
    static create<T>(sourceObj: EventTargetLike, eventName: string, selector?: (...args: Array<any>) => T): Observable<T>;
    constructor(sourceObj: EventTargetLike, eventName: string, selector?: (...args: Array<any>) => T);
    private static setupSubscription<T>(sourceObj, eventName, handler, subscriber);
    protected _subscribe(subscriber: Subscriber<T>): void;
}
