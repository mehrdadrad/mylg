import { Type } from '../type';
import { PlatformReflectionCapabilities } from './platform_reflection_capabilities';
import { GetterFn, MethodFn, SetterFn } from './types';
export declare class ReflectionCapabilities implements PlatformReflectionCapabilities {
    private _reflect;
    constructor(reflect?: any);
    isReflectionEnabled(): boolean;
    factory(t: Type<any>): Function;
    parameters(typeOrFunc: Type<any>): any[][];
    annotations(typeOrFunc: Type<any>): any[];
    propMetadata(typeOrFunc: any): {
        [key: string]: any[];
    };
    interfaces(type: Type<any>): any[];
    hasLifecycleHook(type: any, lcInterface: Type<any>, lcProperty: string): boolean;
    getter(name: string): GetterFn;
    setter(name: string): SetterFn;
    method(name: string): MethodFn;
    importUri(type: any): string;
    resolveIdentifier(name: string, moduleUrl: string, runtime: any): any;
    resolveEnum(enumIdentifier: any, name: string): any;
}
