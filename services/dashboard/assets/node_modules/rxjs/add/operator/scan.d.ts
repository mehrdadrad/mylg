import { ScanSignature } from '../../operator/scan';
declare module '../../Observable' {
    interface Observable<T> {
        scan: ScanSignature<T>;
    }
}
