declare const theme: {
    base: string;
    inherit: boolean;
    rules: Array<{
        foreground?: string;
        background?: string;
        token: string;
        fontStyle?: string;
    }>;
    colors: Record<string, string>;
};

export default theme;
