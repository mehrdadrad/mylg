import { PlatformRef, Provider } from '@angular/core';
export * from './private_export';
/**
 * @experimental
 */
export declare const RESOURCE_CACHE_PROVIDER: Provider[];
/**
 * @experimental API related to bootstrapping are still under review.
 */
export declare const platformBrowserDynamic: (extraProviders?: Provider[]) => PlatformRef;
