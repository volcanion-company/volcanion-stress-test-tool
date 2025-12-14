import { ReactNode } from 'react'
import clsx from 'clsx'

interface CardProps {
  children: ReactNode
  className?: string
  title?: string
  action?: ReactNode
}

export default function Card({ children, className, title, action }: CardProps) {
  return (
    <div className={clsx('bg-white dark:bg-gray-800 rounded-lg shadow', className)}>
      {(title || action) && (
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          {title && (
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              {title}
            </h3>
          )}
          {action}
        </div>
      )}
      <div className="p-6">
        {children}
      </div>
    </div>
  )
}
