import { combineLatestStatic } from '../../operator/combineLatest';
declare module '../../Observable' {
    namespace Observable {
        let combineLatest: typeof combineLatestStatic;
    }
}
