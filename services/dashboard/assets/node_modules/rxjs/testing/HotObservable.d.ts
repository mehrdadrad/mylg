import { Subject } from '../Subject';
import { Subscriber } from '../Subscriber';
import { TeardownLogic } from '../Subscription';
import { Scheduler } from '../Scheduler';
import { TestMessage } from './TestMessage';
import { SubscriptionLog } from './SubscriptionLog';
import { SubscriptionLoggable } from './SubscriptionLoggable';
/**
 * We need this JSDoc comment for affecting ESDoc.
 * @ignore
 * @extends {Ignored}
 */
export declare class HotObservable<T> extends Subject<T> implements SubscriptionLoggable {
    messages: TestMessage[];
    subscriptions: SubscriptionLog[];
    scheduler: Scheduler;
    logSubscribedFrame: () => number;
    logUnsubscribedFrame: (index: number) => void;
    constructor(messages: TestMessage[], scheduler: Scheduler);
    protected _subscribe(subscriber: Subscriber<any>): TeardownLogic;
    setup(): void;
}
