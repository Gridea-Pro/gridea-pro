export { };

declare global {
    interface Window {
        go: any;
        runtime: any;
        [key: string]: any;
    }
}
