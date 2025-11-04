/**
 * Simple URL Builder that helps construct URLs with query parameters
 * Works with any base URL (with or without protocol) - doesn't modify the base URL
 */
export class URLBuilder {
    private baseUrl: string;
    private searchParams: URLSearchParams;

    constructor(baseUrl: string) {
        // Use the base URL as-is, just remove trailing slash if present
        this.baseUrl = baseUrl.replace(/\/$/, '');
        this.searchParams = new URLSearchParams();
    }

    /**
     * Appends a path to the base URL
     * @param path - The path to append (with or without leading slash)
     * @returns URLBuilder instance for chaining
     */
    path(path: string): URLBuilder {
        // Ensure path starts with /
        const normalizedPath = path.startsWith('/') ? path : `/${path}`;
        this.baseUrl += normalizedPath;
        return this;
    }

    /**
     * Adds a query parameter
     * @param key - Parameter key
     * @param value - Parameter value
     * @returns URLBuilder instance for chaining
     */
    param(key: string, value: string | number | boolean): URLBuilder {
        this.searchParams.append(key, value.toString());
        return this;
    }

    /**
     * Adds multiple query parameters from an object
     * @param params - Object containing key-value pairs
     * @returns URLBuilder instance for chaining
     */
    params(params: Record<string, string | number | boolean | undefined | null>): URLBuilder {
        Object.entries(params).forEach(([key, value]) => {
            if (value !== undefined && value !== null && value !== '') {
                this.searchParams.append(key, value.toString());
            }
        });
        return this;
    }

    /**
     * Clears all query parameters
     * @returns URLBuilder instance for chaining
     */
    clearParams(): URLBuilder {
        this.searchParams = new URLSearchParams();
        return this;
    }

    /**
     * Builds and returns the final URL string
     * @returns Complete URL string
     */
    build(): string {
        const queryString = this.searchParams.toString();
        return queryString ? `${this.baseUrl}?${queryString}` : this.baseUrl;
    }

    /**
     * Returns the URL as a string (alias for build())
     * @returns Complete URL string
     */
    toString(): string {
        return this.build();
    }

    /**
     * Static factory method to create a new URLBuilder instance
     * @param baseUrl - The base URL
     * @returns New URLBuilder instance
     */
    static from(baseUrl: string): URLBuilder {
        return new URLBuilder(baseUrl);
    }

    /**
     * Static method to quickly build a URL with path and params
     * @param baseUrl - The base URL
     * @param path - Optional path to append
     * @param params - Optional query parameters
     * @returns Complete URL string
     */
    static build(
        baseUrl: string, 
        path?: string, 
        params?: Record<string, string | number | boolean | undefined | null>
    ): string {
        const builder = new URLBuilder(baseUrl);
        
        if (path) {
            builder.path(path);
        }
        
        if (params) {
            builder.params(params);
        }
        
        return builder.build();
    }
}

export default URLBuilder;