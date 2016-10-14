import { CompileNgModuleMetadata, StaticSymbol } from './compile_metadata';
import { DirectiveNormalizer } from './directive_normalizer';
import { CompileMetadataResolver } from './metadata_resolver';
import { NgModuleCompiler } from './ng_module_compiler';
import { OutputEmitter } from './output/abstract_emitter';
import { StyleCompiler } from './style_compiler';
import { TemplateParser } from './template_parser/template_parser';
import { ViewCompiler } from './view_compiler/view_compiler';
export declare class SourceModule {
    moduleUrl: string;
    source: string;
    constructor(moduleUrl: string, source: string);
}
export declare class NgModulesSummary {
    ngModuleByComponent: Map<StaticSymbol, CompileNgModuleMetadata>;
    constructor(ngModuleByComponent: Map<StaticSymbol, CompileNgModuleMetadata>);
}
export declare class OfflineCompiler {
    private _metadataResolver;
    private _directiveNormalizer;
    private _templateParser;
    private _styleCompiler;
    private _viewCompiler;
    private _ngModuleCompiler;
    private _outputEmitter;
    private _localeId;
    private _translationFormat;
    private _animationParser;
    private _animationCompiler;
    constructor(_metadataResolver: CompileMetadataResolver, _directiveNormalizer: DirectiveNormalizer, _templateParser: TemplateParser, _styleCompiler: StyleCompiler, _viewCompiler: ViewCompiler, _ngModuleCompiler: NgModuleCompiler, _outputEmitter: OutputEmitter, _localeId: string, _translationFormat: string);
    analyzeModules(ngModules: StaticSymbol[]): NgModulesSummary;
    clearCache(): void;
    compile(moduleUrl: string, ngModulesSummary: NgModulesSummary, components: StaticSymbol[], ngModules: StaticSymbol[]): Promise<SourceModule[]>;
    private _compileModule(ngModuleType, targetStatements);
    private _compileComponentFactory(compMeta, fileSuffix, targetStatements);
    private _compileComponent(compMeta, directives, pipes, schemas, componentStyles, fileSuffix, targetStatements);
    private _codgenStyles(stylesCompileResult, fileSuffix);
    private _codegenSourceModule(moduleUrl, statements, exportedVars);
}
