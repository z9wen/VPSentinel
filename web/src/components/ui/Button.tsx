import type { ButtonHTMLAttributes, ReactNode } from 'react'

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger'

const variants: Record<Variant, string> = {
  primary: 'btn-primary',
  secondary: 'border-base-300 bg-base-100 hover:bg-base-200',
  ghost: 'btn-ghost',
  danger: 'btn-error',
}

type Props = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: Variant
  size?: 'sm' | 'md'
  children: ReactNode
}

export function Button({ variant = 'secondary', size = 'md', className = '', children, ...props }: Props) {
  const sizing = size === 'sm' ? 'btn-sm' : ''
  return (
    <button className={`btn gap-2 font-medium shadow-none ${variants[variant]} ${sizing} ${className}`} {...props}>
      {children}
    </button>
  )
}
