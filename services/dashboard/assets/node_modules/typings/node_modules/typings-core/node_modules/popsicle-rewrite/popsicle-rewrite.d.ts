declare function popsicleRewrite (rewrites: { [s: string]:string }): (req: any, next: () => any) => any;

export = popsicleRewrite;
