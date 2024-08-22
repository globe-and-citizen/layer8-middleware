export declare function tunnel(req: any, res: any, next: any): void;
export declare function _static(dir: any): (req: any, res: any, next: any) => void;
export { _static as static };
export declare function multipart(options: any): {
    single: (name: any) => (req: any, res: any, next: any) => void;
    array: (name: any) => (req: any, res: any, next: any) => void;
};
