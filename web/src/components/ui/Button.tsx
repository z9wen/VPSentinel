import type { ButtonHTMLAttributes, ReactNode } from 'react'

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger'

const variants: Record<Variant, string> = {
  primary: 'btn-primary',
  secondary: 'btn-outline border-base-300 bg-base-200',
  ghost: 'btn-ghost',
  danger: 'btn-ghost text-error hover:bg-error/10',
}

type Props = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: Variant
  size?: 'sm' | 'md'
  children: ReactNode
}

export function Button({ variant = 'secondary', size = 'md', className = '', children, ...props }: Props) {
  return <button className={`btn ${variants[variant]} ${size === 'sm' ? 'btn-sm' : ''} ${className}`} {...props}>{children}</button>
}
