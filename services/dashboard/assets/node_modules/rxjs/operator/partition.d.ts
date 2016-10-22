import { Observable } from '../Observable';
/**
 * @param predicate
 * @param thisArg
 * @return {Observable<T>[]}
 * @method partition
 * @owner Observable
 */
export declare function partition<T>(predicate: (value: T) => boolean, thisArg?: any): [Observable<T>, Observable<T>];
export interface PartitionSignature<T> {
    (predicate: (value: T) => boolean, thisArg?: any): [Observable<T>, Observable<T>];
}
