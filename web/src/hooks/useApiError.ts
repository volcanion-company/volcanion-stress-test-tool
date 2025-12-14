import { useState, useCallback } from 'react';
import { AxiosError } from 'axios';

export interface ApiError {
  message: string;
  code?: string;
  status?: number;
  details?: Record<string, string[]>;
  retryable: boolean;
}

interface UseApiErrorReturn {
  error: ApiError | null;
  setError: (error: ApiError | null) => void;
  clearError: () => void;
  handleError: (err: unknown) => ApiError;
  isRetrying: boolean;
  retry: (fn: () => Promise<void>) => Promise<void>;
}

export function parseApiError(err: unknown): ApiError {
  if (err instanceof AxiosError) {
    const status = err.response?.status;
    const data = err.response?.data;

    // Network error or timeout
    if (!err.response) {
      return {
        message: 'Unable to connect to server. Please check your internet connection.',
        code: 'NETWORK_ERROR',
        retryable: true,
      };
    }

    // Server provided error message
    if (data?.message) {
      return {
        message: data.message,
        code: data.code || `HTTP_${status}`,
        status,
        details: data.errors,
        retryable: status === 503 || status === 429 || Boolean(status && status >= 500),
      };
    }

    // Standard HTTP errors
    switch (status) {
      case 400:
        return {
          message: 'Invalid request. Please check your input.',
          code: 'BAD_REQUEST',
          status,
          retryable: false,
        };
      case 401:
        return {
          message: 'Your session has expired. Please log in again.',
          code: 'UNAUTHORIZED',
          status,
          retryable: false,
        };
      case 403:
        return {
          message: 'You do not have permission to perform this action.',
          code: 'FORBIDDEN',
          status,
          retryable: false,
        };
      case 404:
        return {
          message: 'The requested resource was not found.',
          code: 'NOT_FOUND',
          status,
          retryable: false,
        };
      case 409:
        return {
          message: 'A conflict occurred. The resource may have been modified.',
          code: 'CONFLICT',
          status,
          retryable: true,
        };
      case 422:
        return {
          message: 'Validation failed. Please check your input.',
          code: 'VALIDATION_ERROR',
          status,
          details: data?.errors,
          retryable: false,
        };
      case 429:
        return {
          message: 'Too many requests. Please wait a moment and try again.',
          code: 'RATE_LIMITED',
          status,
          retryable: true,
        };
      case 500:
        return {
          message: 'An internal server error occurred. Please try again later.',
          code: 'SERVER_ERROR',
          status,
          retryable: true,
        };
      case 502:
      case 503:
      case 504:
        return {
          message: 'The server is temporarily unavailable. Please try again later.',
          code: 'SERVICE_UNAVAILABLE',
          status,
          retryable: true,
        };
      default:
        return {
          message: 'An unexpected error occurred.',
          code: `HTTP_${status}`,
          status,
          retryable: Boolean(status && status >= 500),
        };
    }
  }

  // Generic Error
  if (err instanceof Error) {
    return {
      message: err.message || 'An unexpected error occurred.',
      code: 'UNKNOWN_ERROR',
      retryable: false,
    };
  }

  return {
    message: 'An unexpected error occurred.',
    code: 'UNKNOWN_ERROR',
    retryable: false,
  };
}

export function useApiError(): UseApiErrorReturn {
  const [error, setError] = useState<ApiError | null>(null);
  const [isRetrying, setIsRetrying] = useState(false);

  const clearError = useCallback(() => setError(null), []);

  const handleError = useCallback((err: unknown): ApiError => {
    const apiError = parseApiError(err);
    setError(apiError);
    return apiError;
  }, []);

  const retry = useCallback(async (fn: () => Promise<void>): Promise<void> => {
    setIsRetrying(true);
    setError(null);
    try {
      await fn();
    } catch (err) {
      handleError(err);
    } finally {
      setIsRetrying(false);
    }
  }, [handleError]);

  return {
    error,
    setError,
    clearError,
    handleError,
    isRetrying,
    retry,
  };
}

export default useApiError;
