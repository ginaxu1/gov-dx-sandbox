# Consent Portal

A React-based web application that provides a user interface for data owners to approve or reject consent requests. The portal integrates with the Consent Engine to manage the complete consent workflow including OTP verification.

## Overview

- **Technology**: React + TypeScript + Vite
- **Port**: 5173 (default)
- **Purpose**: User interface for consent management
- **Role**: Displays consent information and handles user decisions

## Quick Start

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Features

- **Consent Display**: Shows consent details including requested data fields and purpose
- **Decision Interface**: Allows users to approve or reject consent requests
- **OTP Verification**: Handles OTP verification flow for approved consents
- **Retry Logic**: Supports OTP retry with attempt tracking
- **Status Tracking**: Displays current consent status and progress
- **Error Handling**: Comprehensive error handling and user feedback

## Workflow Integration

### 1. Consent Request Flow

1. **URL Parameter**: Portal receives `consent_id` via URL parameter
2. **Data Fetching**: Fetches consent details from Consent Engine
3. **Display**: Shows consent information to user
4. **Decision**: User can approve or reject the consent

### 2. Approval Flow

1. **User Approval**: User clicks "Approve" button
2. **Status Update**: Calls `PUT /consents/:consentId` to update status to "approved"
3. **OTP Request**: Automatically triggers OTP verification
4. **OTP Input**: User enters OTP code (hardcoded to "123456" for testing)
5. **Verification**: Calls `POST /consents/:consentId/otp` to verify OTP
6. **Success**: Redirects to original application with success message

### 3. Rejection Flow

1. **User Rejection**: User clicks "Reject" button
2. **Status Update**: Calls `PUT /consents/:consentId` to update status to "rejected"
3. **Confirmation**: Shows rejection confirmation message
4. **Redirect**: Redirects to original application with denial message

## API Integration

### Consent Engine Endpoints

| Endpoint | Method | Purpose | Usage |
|----------|--------|---------|-------|
| `/consents/{id}` | GET | Get consent details | Fetch consent information for display |
| `/consents/{id}` | PUT | Update consent status | Submit user decision (approve/reject) |
| `/consents/{id}/otp` | POST | Verify OTP | Verify OTP code for approved consents |

### Request/Response Examples

#### Get Consent Details
```bash
GET /consents/consent_03c134ae
```

Response:
```json
{
  "consent_uuid": "consent_03c134ae",
  "owner_id": "199512345678",
  "data_consumer": "passport-app",
  "status": "pending",
  "type": "realtime",
  "created_at": "2025-09-14T07:28:34+05:30",
  "expires_at": "2025-10-14T07:28:34+05:30",
  "fields": ["personInfo.permanentAddress"],
  "session_id": "session_123",
  "redirect_url": "http://localhost:5173/?consent_id=consent_03c134ae",
  "purpose": "passport_application",
  "message": "Consent required. Please visit the consent portal."
}
```

#### Update Consent Status
```bash
PUT /consents/consent_03c134ae
Content-Type: application/json

{
  "status": "approved",
  "owner_id": "199512345678",
  "message": "Approved via consent portal"
}
```

#### Verify OTP
```bash
POST /consents/consent_03c134ae/otp
Content-Type: application/json

{
  "otp_code": "123456"
}
```

## Testing

### Manual Testing

1. **Start Consent Engine**:
   ```bash
   cd /Users/tmp/gov-dx-sandbox/exchange/consent-engine
   go run .
   ```

2. **Start Consent Portal**:
   ```bash
   cd /Users/tmp/gov-dx-sandbox/portals/consent-portal
   npm run dev
   ```

3. **Test with Default Consent**:
   - Navigate to: `http://localhost:5173/?consent_id=consent_03c134ae`
   - This uses the default hardcoded consent record

4. **Test with New Consent**:
   ```bash
   # Create a new consent
   curl -X POST http://localhost:8081/consents \
     -H "Content-Type: application/json" \
     -d '{
       "app_id": "passport-app",
       "data_fields": [
         {
           "owner_type": "citizen",
           "owner_id": "1991111111",
           "fields": ["personInfo.permanentAddress"]
         }
       ],
       "purpose": "passport_application",
       "session_id": "session_test",
       "redirect_url": "https://passport-app.gov.lk"
     }'
   
   # Use the consent_id from the response
   # Navigate to: http://localhost:5173/?consent_id={consent_id}
   ```

### Test Scenarios

#### 1. Approval Flow
1. Open portal with valid consent_id
2. Click "Approve" button
3. Enter OTP: "123456"
4. Verify success message and redirect

#### 2. Rejection Flow
1. Open portal with valid consent_id
2. Click "Reject" button
3. Verify rejection message and redirect

#### 3. OTP Retry Flow
1. Open portal with valid consent_id
2. Click "Approve" button
3. Enter wrong OTP (e.g., "000000")
4. Verify retry message
5. Enter correct OTP: "123456"
6. Verify success

#### 4. Error Handling
1. Open portal with invalid consent_id
2. Verify error message
3. Test network error scenarios

## Configuration

### Environment Variables

Create a `.env` file in the project root:

```env
VITE_CONSENT_ENGINE_URL=http://localhost:8081
VITE_CONSENT_ENGINE_PATH=/consents
```

### Default Configuration

- **Consent Engine URL**: `http://localhost:8081`
- **API Path**: `/consents`
- **OTP Code**: `123456` (hardcoded for testing)
- **Max OTP Attempts**: 3

## Component Structure

```
src/
├── App.tsx                 # Main application component
├── components/
│   ├── ConsentDisplay.tsx  # Consent information display
│   ├── DecisionButtons.tsx # Approve/Reject buttons
│   ├── OTPInput.tsx        # OTP verification input
│   └── StatusMessage.tsx   # Status and error messages
├── types/
│   └── consent.ts          # TypeScript interfaces
└── utils/
    └── api.ts              # API utility functions
```

## State Management

The application uses React hooks for state management:

- **Consent Data**: Fetched from Consent Engine API
- **User Decision**: Tracks approve/reject state
- **OTP Flow**: Manages OTP input and verification
- **Error States**: Handles various error conditions
- **Loading States**: Shows loading indicators during API calls

## Error Handling

### API Errors
- **404**: Consent not found
- **400**: Invalid request data
- **500**: Server errors
- **Network**: Connection issues

### User Experience
- Clear error messages
- Retry mechanisms
- Fallback states
- Loading indicators

## Development

### Adding New Features

1. **Update Types**: Add new interfaces in `types/consent.ts`
2. **API Functions**: Add new API calls in `utils/api.ts`
3. **Components**: Create new components as needed
4. **State Management**: Update state handling in `App.tsx`
5. **Testing**: Add test scenarios

### Debugging

1. **Browser DevTools**: Check network requests and console logs
2. **React DevTools**: Inspect component state and props
3. **API Testing**: Use curl or Postman to test Consent Engine directly

## Integration with Other Services

### Orchestration Engine
- Receives consent requests and redirects to portal
- Handles consent status updates
- Manages data access based on consent decisions

### Consent Engine
- Provides consent data and status
- Handles OTP verification
- Manages consent lifecycle

## Security Considerations

- **URL Parameters**: Consent ID passed via URL (consider security implications)
- **OTP Handling**: OTP codes are not stored in browser
- **Data Validation**: All user inputs are validated
- **Error Messages**: Avoid exposing sensitive information in error messages

## Performance

- **Lazy Loading**: Components loaded as needed
- **API Caching**: Consent data cached during session
- **Error Boundaries**: Graceful error handling
- **Loading States**: User feedback during API calls

## Troubleshooting

### Common Issues

1. **CORS Errors**: Ensure Consent Engine allows cross-origin requests
2. **API Connection**: Check Consent Engine is running on correct port
3. **Invalid Consent ID**: Verify consent exists in Consent Engine
4. **OTP Issues**: Check OTP code is correct ("123456" for testing)

### Debug Steps

1. Check browser console for errors
2. Verify API endpoints are accessible
3. Test with default consent ID: `consent_03c134ae`
4. Check network requests in DevTools
5. Verify Consent Engine logs