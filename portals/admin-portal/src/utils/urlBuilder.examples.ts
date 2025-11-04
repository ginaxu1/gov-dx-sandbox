/**
 * URLBuilder Usage Examples
 * 
 * This file demonstrates how to use the URLBuilder class for different scenarios,
 * especially when working with Choreo service URLs that don't have protocol prefixes.
 */

import { URLBuilder } from './urlBuilder';

// Example 1: Basic usage with Choreo service URL (no protocol)
const choreoServiceUrl = 'api-12345.choreoapis.dev';
const simpleUrl = URLBuilder.build(choreoServiceUrl, '/users');
console.log('Simple URL:', simpleUrl);
// Output: https://api-12345.choreoapis.dev/users

// Example 2: Using method chaining
const complexUrl = URLBuilder.from('my-service.choreoapis.dev')
    .path('/api/v1/logs')
    .param('limit', 50)
    .param('status', 'success')
    .param('startDate', '2023-01-01')
    .build();
console.log('Complex URL:', complexUrl);
// Output: https://my-service.choreoapis.dev/api/v1/logs?limit=50&status=success&startDate=2023-01-01

// Example 3: Using with parameters object (like in your logService)
const params = {
    consumerId: 'consumer-123',
    providerId: 'provider-456',
    startDate: '2023-01-01',
    endDate: '2023-12-31',
    limit: 100,
    offset: 0
};

const urlWithParams = URLBuilder.build('logs-service.choreoapis.dev', '/logs', params);
console.log('URL with params:', urlWithParams);
// Output: https://logs-service.choreoapis.dev/logs?consumerId=consumer-123&providerId=provider-456&startDate=2023-01-01&endDate=2023-12-31&limit=100&offset=0

// Example 4: Handling localhost (will use http instead of https)
const localUrl = URLBuilder.build('localhost:3000', '/api/logs');
console.log('Local URL:', localUrl);
// Output: http://localhost:3000/api/logs

// Example 5: URL already has protocol (will not modify)
const existingProtocolUrl = URLBuilder.build('https://api.example.com', '/users');
console.log('Existing protocol URL:', existingProtocolUrl);
// Output: https://api.example.com/users

// Example 6: Filtering undefined/null/empty parameters
const paramsWithEmpty = {
    consumerId: 'consumer-123',
    providerId: '', // Will be filtered out
    startDate: undefined, // Will be filtered out
    endDate: null, // Will be filtered out
    limit: 50
};

const filteredUrl = URLBuilder.build('api.choreoapis.dev', '/logs', paramsWithEmpty);
console.log('Filtered URL:', filteredUrl);
// Output: https://api.choreoapis.dev/logs?consumerId=consumer-123&limit=50

export {};