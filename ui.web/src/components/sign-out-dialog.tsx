import { useLocation } from '@tanstack/react-router'
import { ConfirmDialog } from '@/components/confirm-dialog'

interface SignOutDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function SignOutDialog({ open, onOpenChange }: SignOutDialogProps) {
  const location = useLocation()

  const handleSignOut = () => {
    // Preserve current location for optional post-login return after a clean sign-out.
    const currentPath = encodeURIComponent(location.href)
    window.location.replace(`/sign-out?redirect=${currentPath}`)
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title='Sign out'
      desc='Are you sure you want to sign out? You will need to sign in again to access your account.'
      confirmText='Sign out'
      destructive
      handleConfirm={handleSignOut}
      className='sm:max-w-sm'
    />
  )
}
