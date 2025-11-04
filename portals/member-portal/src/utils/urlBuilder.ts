/**
 * Custom URL Builder that handles URLs without protocol prefixes
 * Useful for Choreo service URLs that don't include http/https
 */
export class URLBuilder {
    private baseUrl: string;
    private searchParams: URLSearchParams;

    constructor(baseUrl: string) {
        // Normalize the base URL - add https:// if no protocol is present
        this.baseUrl = this.normalizeBaseUrl(baseUrl);
        this.searchParams = new URLSearchParams();
    }

    /**
     * Normalizes the base URL by adding protocol if missing
     * @param url - The base URL which may or may not have a protocol
     * @returns Normalized URL with protocol
     */
    private normalizeBaseUrl(url: string): string {
        // Remove trailing slash if present
        const cleanUrl = url.replace(/\/$/, '');
        
        // Check if URL already has a protocol
        if (cleanUrl.match(/^https?:\/\//)) {
            return cleanUrl;
        }
        
        // Check if it's a localhost or development URL
        if (cleanUrl.includes('localhost') || cleanUrl.includes('127.0.0.1')) {
            return `http://${cleanUrl}`;
        }
        
        // For production URLs (including Choreo service URLs), default to https
        return `https://${cleanUrl}`;
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