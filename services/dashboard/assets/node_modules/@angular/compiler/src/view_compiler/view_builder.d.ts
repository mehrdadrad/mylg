import { CompileIdentifierMetadata } from '../compile_metadata';
import * as o from '../output/output_ast';
import { TemplateAst } from '../template_parser/template_ast';
import { CompileView } from './compile_view';
export declare class ViewFactoryDependency {
    comp: CompileIdentifierMetadata;
    placeholder: CompileIdentifierMetadata;
    constructor(comp: CompileIdentifierMetadata, placeholder: CompileIdentifierMetadata);
}
export declare class ComponentFactoryDependency {
    comp: CompileIdentifierMetadata;
    placeholder: CompileIdentifierMetadata;
    constructor(comp: CompileIdentifierMetadata, placeholder: CompileIdentifierMetadata);
}
export declare function buildView(view: CompileView, template: TemplateAst[], targetDependencies: Array<ViewFactoryDependency | ComponentFactoryDependency>): number;
export declare function finishView(view: CompileView, targetStatements: o.Statement[]): void;
