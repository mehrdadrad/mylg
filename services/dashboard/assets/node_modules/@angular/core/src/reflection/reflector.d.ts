import { Type } from '../type';
import { PlatformReflectionCapabilities } from './platform_reflection_capabilities';
import { ReflectorReader } from './reflector_reader';
import { GetterFn, MethodFn, SetterFn } from './types';
export { PlatformReflectionCapabilities } from './platform_reflection_capabilities';
export { GetterFn, MethodFn, SetterFn } from './types';
/**
 * Reflective information about a symbol, including annotations, interfaces, and other metadata.
 */
export declare class ReflectionInfo {
    annotations: any[];
    parameters: any[][];
    factory: Function;
    interfaces: any[];
    propMetadata: {
        [key: string]: any[];
    };
    constructor(annotations?: any[], parameters?: any[][], factory?: Function, interfaces?: any[], propMetadata?: {
        [key: string]: any[];
    });
}
/**
 * Provides access to reflection data about symbols. Used internally by Angular
 * to power dependency injection and compilation.
 */
export declare class Reflector extends ReflectorReader {
    reflectionCapabilities: PlatformReflectionCapabilities;
    constructor(reflectionCapabilities: PlatformReflectionCapabilities);
    updateCapabilities(caps: PlatformReflectionCapabilities): void;
    isReflectionEnabled(): boolean;
    /**
     * Causes `this` reflector to track keys used to access
     * {@link ReflectionInfo} objects.
     */
    trackUsage(): void;
    /**
     * Lists types for which reflection information was not requested since
     * {@link #trackUsage} was called. This list could later be audited as
     * potential dead code.
     */
    listUnusedKeys(): any[];
    registerFunction(func: Function, funcInfo: ReflectionInfo): void;
    registerType(type: Type<any>, typeInfo: ReflectionInfo): void;
    registerGetters(getters: {
        [key: string]: GetterFn;
    }): void;
    registerSetters(setters: {
        [key: string]: SetterFn;
    }): void;
    registerMethods(methods: {
        [key: string]: MethodFn;
    }): void;
    factory(type: Type<any>): Function;
    parameters(typeOrFunc: Type<any>): any[][];
    annotations(typeOrFunc: Type<any>): any[];
    propMetadata(typeOrFunc: Type<any>): {
        [key: string]: any[];
    };
    interfaces(type: Type<any>): any[];
    hasLifecycleHook(type: any, lcInterface: Type<any>, lcProperty: string): boolean;
    getter(name: string): GetterFn;
    setter(name: string): SetterFn;
    method(name: string): MethodFn;
    importUri(type: any): string;
    resolveIdentifier(name: string, moduleUrl: string, runtime: any): any;
    resolveEnum(identifier: any, name: string): any;
}
