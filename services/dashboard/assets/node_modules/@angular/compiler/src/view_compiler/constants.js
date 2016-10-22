/**
 * @license
 * Copyright Google Inc. All Rights Reserved.
 *
 * Use of this source code is governed by an MIT-style license that can be
 * found in the LICENSE file at https://angular.io/license
 */
import { ChangeDetectionStrategy, ViewEncapsulation } from '@angular/core';
import { Identifiers, resolveEnumIdentifier, resolveIdentifier } from '../identifiers';
import * as o from '../output/output_ast';
import { ChangeDetectorStatus, ViewType } from '../private_import_core';
function _enumExpression(classIdentifier, name) {
    return o.importExpr(resolveEnumIdentifier(classIdentifier, name));
}
export var ViewTypeEnum = (function () {
    function ViewTypeEnum() {
    }
    ViewTypeEnum.fromValue = function (value) {
        var viewType = resolveIdentifier(Identifiers.ViewType);
        switch (value) {
            case ViewType.HOST:
                return _enumExpression(viewType, 'HOST');
            case ViewType.COMPONENT:
                return _enumExpression(viewType, 'COMPONENT');
            case ViewType.EMBEDDED:
                return _enumExpression(viewType, 'EMBEDDED');
            default:
                throw Error("Inavlid ViewType value: " + value);
        }
    };
    return ViewTypeEnum;
}());
export var ViewEncapsulationEnum = (function () {
    function ViewEncapsulationEnum() {
    }
    ViewEncapsulationEnum.fromValue = function (value) {
        var viewEncapsulation = resolveIdentifier(Identifiers.ViewEncapsulation);
        switch (value) {
            case ViewEncapsulation.Emulated:
                return _enumExpression(viewEncapsulation, 'Emulated');
            case ViewEncapsulation.Native:
                return _enumExpression(viewEncapsulation, 'Native');
            case ViewEncapsulation.None:
                return _enumExpression(viewEncapsulation, 'None');
            default:
                throw Error("Inavlid ViewEncapsulation value: " + value);
        }
    };
    return ViewEncapsulationEnum;
}());
export var ChangeDetectionStrategyEnum = (function () {
    function ChangeDetectionStrategyEnum() {
    }
    ChangeDetectionStrategyEnum.fromValue = function (value) {
        var changeDetectionStrategy = resolveIdentifier(Identifiers.ChangeDetectionStrategy);
        switch (value) {
            case ChangeDetectionStrategy.OnPush:
                return _enumExpression(changeDetectionStrategy, 'OnPush');
            case ChangeDetectionStrategy.Default:
                return _enumExpression(changeDetectionStrategy, 'Default');
            default:
                throw Error("Inavlid ChangeDetectionStrategy value: " + value);
        }
    };
    return ChangeDetectionStrategyEnum;
}());
export var ChangeDetectorStatusEnum = (function () {
    function ChangeDetectorStatusEnum() {
    }
    ChangeDetectorStatusEnum.fromValue = function (value) {
        var changeDetectorStatus = resolveIdentifier(Identifiers.ChangeDetectorStatus);
        switch (value) {
            case ChangeDetectorStatus.CheckOnce:
                return _enumExpression(changeDetectorStatus, 'CheckOnce');
            case ChangeDetectorStatus.Checked:
                return _enumExpression(changeDetectorStatus, 'Checked');
            case ChangeDetectorStatus.CheckAlways:
                return _enumExpression(changeDetectorStatus, 'CheckAlways');
            case ChangeDetectorStatus.Detached:
                return _enumExpression(changeDetectorStatus, 'Detached');
            case ChangeDetectorStatus.Errored:
                return _enumExpression(changeDetectorStatus, 'Errored');
            case ChangeDetectorStatus.Destroyed:
                return _enumExpression(changeDetectorStatus, 'Destroyed');
            default:
                throw Error("Inavlid ChangeDetectorStatus value: " + value);
        }
    };
    return ChangeDetectorStatusEnum;
}());
export var ViewConstructorVars = (function () {
    function ViewConstructorVars() {
    }
    ViewConstructorVars.viewUtils = o.variable('viewUtils');
    ViewConstructorVars.parentInjector = o.variable('parentInjector');
    ViewConstructorVars.declarationEl = o.variable('declarationEl');
    return ViewConstructorVars;
}());
export var ViewProperties = (function () {
    function ViewProperties() {
    }
    ViewProperties.renderer = o.THIS_EXPR.prop('renderer');
    ViewProperties.projectableNodes = o.THIS_EXPR.prop('projectableNodes');
    ViewProperties.viewUtils = o.THIS_EXPR.prop('viewUtils');
    return ViewProperties;
}());
export var EventHandlerVars = (function () {
    function EventHandlerVars() {
    }
    EventHandlerVars.event = o.variable('$event');
    return EventHandlerVars;
}());
export var InjectMethodVars = (function () {
    function InjectMethodVars() {
    }
    InjectMethodVars.token = o.variable('token');
    InjectMethodVars.requestNodeIndex = o.variable('requestNodeIndex');
    InjectMethodVars.notFoundResult = o.variable('notFoundResult');
    return InjectMethodVars;
}());
export var DetectChangesVars = (function () {
    function DetectChangesVars() {
    }
    DetectChangesVars.throwOnChange = o.variable("throwOnChange");
    DetectChangesVars.changes = o.variable("changes");
    DetectChangesVars.changed = o.variable("changed");
    DetectChangesVars.valUnwrapper = o.variable("valUnwrapper");
    return DetectChangesVars;
}());
//# sourceMappingURL=constants.js.map