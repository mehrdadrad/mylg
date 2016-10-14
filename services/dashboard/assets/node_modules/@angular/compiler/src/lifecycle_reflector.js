/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
import { AfterContentChecked, AfterContentInit, AfterViewChecked, AfterViewInit, DoCheck, OnChanges, OnDestroy, OnInit } from '@angular/core';
import { MapWrapper } from './facade/collection';
import { LifecycleHooks, reflector } from './private_import_core';
var LIFECYCLE_INTERFACES = MapWrapper.createFromPairs([
    [LifecycleHooks.OnInit, OnInit],
    [LifecycleHooks.OnDestroy, OnDestroy],
    [LifecycleHooks.DoCheck, DoCheck],
    [LifecycleHooks.OnChanges, OnChanges],
    [LifecycleHooks.AfterContentInit, AfterContentInit],
    [LifecycleHooks.AfterContentChecked, AfterContentChecked],
    [LifecycleHooks.AfterViewInit, AfterViewInit],
    [LifecycleHooks.AfterViewChecked, AfterViewChecked],
]);
var LIFECYCLE_PROPS = MapWrapper.createFromPairs([
    [LifecycleHooks.OnInit, 'ngOnInit'],
    [LifecycleHooks.OnDestroy, 'ngOnDestroy'],
    [LifecycleHooks.DoCheck, 'ngDoCheck'],
    [LifecycleHooks.OnChanges, 'ngOnChanges'],
    [LifecycleHooks.AfterContentInit, 'ngAfterContentInit'],
    [LifecycleHooks.AfterContentChecked, 'ngAfterContentChecked'],
    [LifecycleHooks.AfterViewInit, 'ngAfterViewInit'],
    [LifecycleHooks.AfterViewChecked, 'ngAfterViewChecked'],
]);
export function hasLifecycleHook(hook, token) {
    var lcInterface = LIFECYCLE_INTERFACES.get(hook);
    var lcProp = LIFECYCLE_PROPS.get(hook);
    return reflector.hasLifecycleHook(token, lcInterface, lcProp);
}
//# sourceMappingURL=lifecycle_reflector.js.map