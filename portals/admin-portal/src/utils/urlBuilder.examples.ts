/**
 * URLBuilder Usage Examples
 * 
 * This file demonstrates how to use the URLBuilder class for different scenarios,
 * especially when working with Choreo service URLs that are relative paths.
 */

import { URLBuilder } from './urlBuilder';

// Example 1: Basic usage with Choreo service URL (relative path)
const choreoServiceUrl = '/choreo-apis/opendif-ndx/audit-service/v1';
const simpleUrl = URLBuilder.build(choreoServiceUrl, '/logs');
console.log('Simple URL:', simpleUrl);
// Output: /choreo-apis/opendif-ndx/audit-service/v1/logs

// Example 2: Using method chaining with Choreo paths
const complexUrl = URLBuilder.from('/choreo-apis/opendif-ndx/api-server/v1')
    .path('/members')
    .param('limit', 50)
    .param('status', 'active')
    .param('startDate', '2023-01-01')
    .build();
console.log('Complex URL:', complexUrl);
// Output: /choreo-apis/opendif-ndx/api-server/v1/members?limit=50&status=active&startDate=2023-01-01

// Example 3: Using with parameters object (like in your logService)
const params = {
    consumerId: 'consumer-123',
    providerId: 'provider-456',
    startDate: '2023-01-01',
    endDate: '2023-12-31',
    limit: 100,
    offset: 0
};

const urlWithParams = URLBuilder.build('/choreo-apis/opendif-ndx/audit-service/v1', '/logs', params);
console.log('URL with params:', urlWithParams);
// Output: /choreo-apis/opendif-ndx/audit-service/v1/logs?consumerId=consumer-123&providerId=provider-456&startDate=2023-01-01&endDate=2023-12-31&limit=100&offset=0

// Example 4: Handling localhost for development
const localUrl = URLBuilder.build('http://localhost:3000', '/api/logs');
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

const filteredUrl = URLBuilder.build('/choreo-apis/opendif-ndx/audit-service/v1', '/logs', paramsWithEmpty);
console.log('Filtered URL:', filteredUrl);
// Output: /choreo-apis/opendif-ndx/audit-service/v1/logs?consumerId=consumer-123&limit=50

// Example 7: Real-world usage in your services
const apiUrl = '/choreo-apis/opendif-ndx/api-server/v1';
const membersUrl = URLBuilder.build(apiUrl, '/members', { idpUserId: 'user-123' });
console.log('Members URL:', membersUrl);
// Output: /choreo-apis/opendif-ndx/api-server/v1/members?idpUserId=user-123

const schemasUrl = URLBuilder.build(apiUrl, '/schemas', { memberId: 'member-456', status: 'approved' });
console.log('Schemas URL:', schemasUrl);
// Output: /choreo-apis/opendif-ndx/api-server/v1/schemas?memberId=member-456&status=approved

export {};