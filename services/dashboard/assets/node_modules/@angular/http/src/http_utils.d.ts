import { RequestMethod } from './enums';
export declare function normalizeMethodName(method: string | RequestMethod): RequestMethod;
export declare const isSuccess: (status: number) => boolean;
export declare function getResponseURL(xhr: any): string;
export declare function stringToArrayBuffer(input: String): ArrayBuffer;
export { isJsObject } from '../src/facade/lang';
