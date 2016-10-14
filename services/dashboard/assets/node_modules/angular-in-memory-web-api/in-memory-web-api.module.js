"use strict";
var core_1 = require('@angular/core');
var http_1 = require('@angular/http');
var in_memory_backend_service_1 = require('./in-memory-backend.service');
// AoT requires factory to be exported
function inMemoryBackendServiceFactory(injector, dbService, options) {
    var backend = new in_memory_backend_service_1.InMemoryBackendService(injector, dbService, options);
    return backend;
}
exports.inMemoryBackendServiceFactory = inMemoryBackendServiceFactory;
var InMemoryWebApiModule = (function () {
    function InMemoryWebApiModule() {
    }
    /**
    *  Prepare in-memory-web-api in the root/boot application module
    *  with class that implements InMemoryDbService and creates an in-memory database.
    *
    * @param {Type} dbCreator - Class that creates seed data for in-memory database. Must implement InMemoryDbService.
    * @param {InMemoryBackendConfigArgs} [options]
    *
    * @example
    * InMemoryWebApiModule.forRoot(dbCreator);
    * InMemoryWebApiModule.forRoot(dbCreator, {useValue: {delay:600}});
    */
    InMemoryWebApiModule.forRoot = function (dbCreator, options) {
        return {
            ngModule: InMemoryWebApiModule,
            providers: [
                { provide: in_memory_backend_service_1.InMemoryDbService, useClass: dbCreator },
                { provide: in_memory_backend_service_1.InMemoryBackendConfig, useValue: options },
            ]
        };
    };
    InMemoryWebApiModule.decorators = [
        { type: core_1.NgModule, args: [{
                    // Must useFactory for AoT
                    // https://github.com/angular/angular/issues/11178
                    providers: [{ provide: http_1.XHRBackend,
                            useFactory: inMemoryBackendServiceFactory,
                            deps: [core_1.Injector, in_memory_backend_service_1.InMemoryDbService, in_memory_backend_service_1.InMemoryBackendConfig] }]
                },] },
    ];
    /** @nocollapse */
    InMemoryWebApiModule.ctorParameters = [];
    return InMemoryWebApiModule;
}());
exports.InMemoryWebApiModule = InMemoryWebApiModule;
//# sourceMappingURL=in-memory-web-api.module.js.map