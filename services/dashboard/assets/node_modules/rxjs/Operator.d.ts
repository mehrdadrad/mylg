import { Subscriber } from './Subscriber';
export declare class Operator<T, R> {
    call(subscriber: Subscriber<R>, source: any): any;
}
