import { Response } from 'express';

/**
 * Sends a standardized success response
 * @param res The Express response object
 * @param data The payload to be sent in the 'data' field
 * @param message A human-readable success message
 * @param statusCode The HTTP status code (defaults to 200)
 */
export const sendSuccess = (res: Response, data: any, message: string = 'Request successful', statusCode: number = 200) => {
    return res.status(statusCode).json({
        status: 'success',
        message,
        data,
        error: null,
    });
};

/**
 * Sends a standardized error response
 * @param res The Express response object
 * @param message A human-readable error message
 * @param statusCode The HTTP status code (e.g., 400, 404, 500)
 * @param errorDetails Optional structured error details
 */
export const sendError = (res: Response, message: string, statusCode: number, errorDetails: any = null) => {
    return res.status(statusCode).json({
        status: 'error',
        message,
        data: null,
        error: {
            code: `E${statusCode}`,
            details: errorDetails || message,
        },
    });
};
