# Consent Portal Testing Guide

This README explains how to test the consent-portal with the ConsentEngine.

## Start the Consent Engine Backend

1. `cd exchange/consent-engine`
2. `go run engine.go main.go`

## Generate Consent ID

1. Go to the Postman collection: OpenDIF → Consent Engine → POST (OE) Call Consent Engine
2. Send the POST request
3. Copy the `consent_id` from the response

## Start the Consent Portal

1. `cd portals/consent-portal`
2. `npm install`
3. `npm run dev`

## Test the Portal

1. Navigate to: `http://localhost:5174/?consent={consent_uuid}`
    - Note: `consent_uuid` = the ID copied from the Postman response

## Testing Scenarios

You can test the following:
1. Approve or deny the consent
2. Return to the portal - it will show consent already approved or denied