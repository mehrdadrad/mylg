import { Action } from './Action';
import { FutureAction } from './FutureAction';
/**
 * We need this JSDoc comment for affecting ESDoc.
 * @ignore
 * @extends {Ignored}
 */
export declare class AnimationFrameAction<T> extends FutureAction<T> {
    protected _schedule(state?: T, delay?: number): Action<T>;
    protected _unsubscribe(): void;
}
