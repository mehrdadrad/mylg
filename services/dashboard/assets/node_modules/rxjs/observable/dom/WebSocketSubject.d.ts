import { Subject } from '../../Subject';
import { Subscriber } from '../../Subscriber';
import { Observable } from '../../Observable';
import { Operator } from '../../Operator';
import { Subscription } from '../../Subscription';
import { Observer, NextObserver } from '../../Observer';
export interface WebSocketSubjectConfig {
    url: string;
    protocol?: string | Array<string>;
    resultSelector?: <T>(e: MessageEvent) => T;
    openObserver?: NextObserver<Event>;
    closeObserver?: NextObserver<CloseEvent>;
    closingObserver?: NextObserver<void>;
    WebSocketCtor?: {
        new (url: string, protocol?: string | Array<string>): WebSocket;
    };
}
/**
 * We need this JSDoc comment for affecting ESDoc.
 * @extends {Ignored}
 * @hide true
 */
export declare class WebSocketSubject<T> extends Subject<T> {
    url: string;
    protocol: string | Array<string>;
    socket: WebSocket;
    openObserver: NextObserver<Event>;
    closeObserver: NextObserver<CloseEvent>;
    closingObserver: NextObserver<void>;
    WebSocketCtor: {
        new (url: string, protocol?: string | Array<string>): WebSocket;
    };
    resultSelector(e: MessageEvent): any;
    /**
     * @param urlConfigOrSource
     * @return {WebSocketSubject}
     * @static true
     * @name webSocket
     * @owner Observable
     */
    static create<T>(urlConfigOrSource: string | WebSocketSubjectConfig): WebSocketSubject<T>;
    constructor(urlConfigOrSource: string | WebSocketSubjectConfig | Observable<T>, destination?: Observer<T>);
    lift<R>(operator: Operator<T, R>): WebSocketSubject<T>;
    multiplex(subMsg: () => any, unsubMsg: () => any, messageFilter: (value: T) => boolean): Observable<{}>;
    protected _unsubscribe(): void;
    protected _subscribe(subscriber: Subscriber<T>): Subscription;
}
