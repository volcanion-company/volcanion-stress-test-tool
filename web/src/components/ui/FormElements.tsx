import { forwardRef, InputHTMLAttributes, TextareaHTMLAttributes, SelectHTMLAttributes, ReactNode, useId } from 'react';

// Base input styles
const baseInputStyles = `
  w-full px-3 py-2 
  bg-white dark:bg-gray-800 
  border border-gray-300 dark:border-gray-600 
  rounded-lg shadow-sm
  text-gray-900 dark:text-white
  placeholder-gray-400 dark:placeholder-gray-500
  focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
  disabled:bg-gray-100 disabled:dark:bg-gray-900 disabled:cursor-not-allowed
  transition-colors
`;

const errorStyles = 'border-red-500 dark:border-red-500 focus:ring-red-500 focus:border-red-500';

// Label component
interface LabelProps {
  htmlFor: string;
  children: ReactNode;
  required?: boolean;
  className?: string;
}

export function Label({ htmlFor, children, required, className = '' }: LabelProps) {
  return (
    <label
      htmlFor={htmlFor}
      className={`block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1 ${className}`}
    >
      {children}
      {required && <span className="text-red-500 ml-1" aria-hidden="true">*</span>}
    </label>
  );
}

// Helper text component
interface HelperTextProps {
  id?: string;
  children: ReactNode;
  error?: boolean;
}

export function HelperText({ id, children, error }: HelperTextProps) {
  return (
    <p
      id={id}
      className={`mt-1 text-sm ${error ? 'text-red-600 dark:text-red-400' : 'text-gray-500 dark:text-gray-400'}`}
      role={error ? 'alert' : undefined}
    >
      {children}
    </p>
  );
}

// Input component
interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, helperText, className = '', id: providedId, required, ...props }, ref) => {
    const generatedId = useId();
    const id = providedId || generatedId;
    const errorId = `${id}-error`;
    const helperId = `${id}-helper`;
    const hasError = !!error;

    return (
      <div className="w-full">
        {label && (
          <Label htmlFor={id} required={required}>
            {label}
          </Label>
        )}
        <input
          ref={ref}
          id={id}
          className={`${baseInputStyles} ${hasError ? errorStyles : ''} ${className}`}
          aria-invalid={hasError}
          aria-describedby={hasError ? errorId : helperText ? helperId : undefined}
          required={required}
          {...props}
        />
        {error && <HelperText id={errorId} error>{error}</HelperText>}
        {!error && helperText && <HelperText id={helperId}>{helperText}</HelperText>}
      </div>
    );
  }
);

Input.displayName = 'Input';

// Textarea component
interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  error?: string;
  helperText?: string;
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ label, error, helperText, className = '', id: providedId, required, ...props }, ref) => {
    const generatedId = useId();
    const id = providedId || generatedId;
    const errorId = `${id}-error`;
    const helperId = `${id}-helper`;
    const hasError = !!error;

    return (
      <div className="w-full">
        {label && (
          <Label htmlFor={id} required={required}>
            {label}
          </Label>
        )}
        <textarea
          ref={ref}
          id={id}
          className={`${baseInputStyles} ${hasError ? errorStyles : ''} min-h-[100px] ${className}`}
          aria-invalid={hasError}
          aria-describedby={hasError ? errorId : helperText ? helperId : undefined}
          required={required}
          {...props}
        />
        {error && <HelperText id={errorId} error>{error}</HelperText>}
        {!error && helperText && <HelperText id={helperId}>{helperText}</HelperText>}
      </div>
    );
  }
);

Textarea.displayName = 'Textarea';

// Select component
interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  error?: string;
  helperText?: string;
  options: Array<{ value: string; label: string; disabled?: boolean }>;
  placeholder?: string;
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ label, error, helperText, options, placeholder, className = '', id: providedId, required, ...props }, ref) => {
    const generatedId = useId();
    const id = providedId || generatedId;
    const errorId = `${id}-error`;
    const helperId = `${id}-helper`;
    const hasError = !!error;

    return (
      <div className="w-full">
        {label && (
          <Label htmlFor={id} required={required}>
            {label}
          </Label>
        )}
        <select
          ref={ref}
          id={id}
          className={`${baseInputStyles} ${hasError ? errorStyles : ''} ${className}`}
          aria-invalid={hasError}
          aria-describedby={hasError ? errorId : helperText ? helperId : undefined}
          required={required}
          {...props}
        >
          {placeholder && (
            <option value="" disabled>
              {placeholder}
            </option>
          )}
          {options.map((option) => (
            <option key={option.value} value={option.value} disabled={option.disabled}>
              {option.label}
            </option>
          ))}
        </select>
        {error && <HelperText id={errorId} error>{error}</HelperText>}
        {!error && helperText && <HelperText id={helperId}>{helperText}</HelperText>}
      </div>
    );
  }
);

Select.displayName = 'Select';

// Checkbox component
interface CheckboxProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label: string;
  error?: string;
  helperText?: string;
}

export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
  ({ label, error, helperText, className = '', id: providedId, ...props }, ref) => {
    const generatedId = useId();
    const id = providedId || generatedId;
    const errorId = `${id}-error`;
    const helperId = `${id}-helper`;
    const hasError = !!error;

    return (
      <div className="w-full">
        <div className="flex items-center">
          <input
            ref={ref}
            type="checkbox"
            id={id}
            className={`
              h-4 w-4 rounded border-gray-300 dark:border-gray-600
              text-blue-600 
              focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800
              disabled:cursor-not-allowed disabled:opacity-50
              ${hasError ? 'border-red-500' : ''}
              ${className}
            `}
            aria-invalid={hasError}
            aria-describedby={hasError ? errorId : helperText ? helperId : undefined}
            {...props}
          />
          <label
            htmlFor={id}
            className="ml-2 text-sm text-gray-700 dark:text-gray-300 cursor-pointer"
          >
            {label}
          </label>
        </div>
        {error && <HelperText id={errorId} error>{error}</HelperText>}
        {!error && helperText && <HelperText id={helperId}>{helperText}</HelperText>}
      </div>
    );
  }
);

Checkbox.displayName = 'Checkbox';

// Form group component for organizing fields
interface FormGroupProps {
  children: ReactNode;
  className?: string;
}

export function FormGroup({ children, className = '' }: FormGroupProps) {
  return (
    <div className={`space-y-4 ${className}`} role="group">
      {children}
    </div>
  );
}

// Form section with heading
interface FormSectionProps {
  title: string;
  description?: string;
  children: ReactNode;
  className?: string;
}

export function FormSection({ title, description, children, className = '' }: FormSectionProps) {
  return (
    <fieldset className={`border-0 p-0 ${className}`}>
      <legend className="text-lg font-medium text-gray-900 dark:text-white mb-1">
        {title}
      </legend>
      {description && (
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
          {description}
        </p>
      )}
      <div className="space-y-4">
        {children}
      </div>
    </fieldset>
  );
}

export default Input;
