import { useEffect, useState } from 'react';

export interface ApiError {
  message: string;
  code?: string;
  status?: number;
  details?: Record<string, string[]>;
  retryable: boolean;
}

interface ErrorAlertProps {
  error: ApiError | null;
  onDismiss?: () => void;
  onRetry?: () => void;
  className?: string;
  autoDismiss?: boolean;
  autoDismissDelay?: number;
}

export function ErrorAlert({
  error,
  onDismiss,
  onRetry,
  className = '',
  autoDismiss = false,
  autoDismissDelay = 5000,
}: ErrorAlertProps) {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    if (error) {
      setIsVisible(true);
      
      if (autoDismiss && !error.retryable) {
        const timer = setTimeout(() => {
          setIsVisible(false);
          onDismiss?.();
        }, autoDismissDelay);
        return () => clearTimeout(timer);
      }
    } else {
      setIsVisible(false);
    }
  }, [error, autoDismiss, autoDismissDelay, onDismiss]);

  if (!error || !isVisible) return null;

  const handleDismiss = () => {
    setIsVisible(false);
    onDismiss?.();
  };

  return (
    <div
      role="alert"
      aria-live="polite"
      className={`rounded-lg border p-4 ${className} ${
        error.retryable
          ? 'bg-yellow-50 border-yellow-200 dark:bg-yellow-900/20 dark:border-yellow-800'
          : 'bg-red-50 border-red-200 dark:bg-red-900/20 dark:border-red-800'
      }`}
    >
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0">
          {error.retryable ? (
            <svg
              className="w-5 h-5 text-yellow-600 dark:text-yellow-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
          ) : (
            <svg
              className="w-5 h-5 text-red-600 dark:text-red-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          )}
        </div>

        <div className="flex-1 min-w-0">
          <p
            className={`text-sm font-medium ${
              error.retryable
                ? 'text-yellow-800 dark:text-yellow-200'
                : 'text-red-800 dark:text-red-200'
            }`}
          >
            {error.message}
          </p>

          {error.details && Object.keys(error.details).length > 0 && (
            <ul className="mt-2 text-sm text-red-700 dark:text-red-300 list-disc list-inside">
              {Object.entries(error.details).map(([field, messages]: [string, string[]]) =>
                messages.map((msg: string, i: number) => (
                  <li key={`${field}-${i}`}>
                    <span className="font-medium">{field}:</span> {msg}
                  </li>
                ))
              )}
            </ul>
          )}

          {(error.retryable && onRetry) && (
            <button
              onClick={onRetry}
              className="mt-2 text-sm font-medium text-yellow-700 dark:text-yellow-300 hover:text-yellow-900 dark:hover:text-yellow-100 underline"
            >
              Try again
            </button>
          )}
        </div>

        {onDismiss && (
          <button
            onClick={handleDismiss}
            className={`flex-shrink-0 p-1 rounded hover:bg-opacity-20 ${
              error.retryable
                ? 'text-yellow-600 dark:text-yellow-400 hover:bg-yellow-600'
                : 'text-red-600 dark:text-red-400 hover:bg-red-600'
            }`}
            aria-label="Dismiss error"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        )}
      </div>
    </div>
  );
}

// Inline error for form fields
interface InlineErrorProps {
  message?: string;
  id?: string;
}

export function InlineError({ message, id }: InlineErrorProps) {
  if (!message) return null;

  return (
    <p
      id={id}
      role="alert"
      className="mt-1 text-sm text-red-600 dark:text-red-400"
    >
      {message}
    </p>
  );
}

// Toast-style error notification
interface ErrorToastProps {
  error: ApiError | null;
  onDismiss: () => void;
}

export function ErrorToast({ error, onDismiss }: ErrorToastProps) {
  const [isExiting, setIsExiting] = useState(false);

  useEffect(() => {
    if (error) {
      setIsExiting(false);
      const timer = setTimeout(() => {
        setIsExiting(true);
        setTimeout(onDismiss, 300);
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [error, onDismiss]);

  if (!error) return null;

  return (
    <div
      role="alert"
      className={`fixed bottom-4 right-4 max-w-sm bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 p-4 transition-all duration-300 ${
        isExiting ? 'opacity-0 translate-y-2' : 'opacity-100 translate-y-0'
      }`}
    >
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0">
          <svg
            className="w-5 h-5 text-red-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </div>
        <div className="flex-1">
          <p className="text-sm font-medium text-gray-900 dark:text-white">
            {error.message}
          </p>
        </div>
        <button
          onClick={onDismiss}
          className="flex-shrink-0 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>
  );
}

export default ErrorAlert;
