import { useState } from 'react'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { AuthLayout } from '../auth-layout'
import { OtpForm } from './components/otp-form'

type ResendState = 'idle' | 'sending' | 'sent'

export function Otp() {
  const [resendState, setResendState] = useState<ResendState>('idle')
  const [resendMessage, setResendMessage] = useState('')

  const handleResend = () => {
    setResendState('sending')
    setResendMessage('Sending a new verification code…')

    window.setTimeout(() => {
      setResendState('sent')
      setResendMessage('A new verification code was sent. Stay on this screen and enter the latest code.')
    }, 700)
  }

  return (
    <AuthLayout>
      <Card className='gap-4'>
        <CardHeader>
          <CardTitle className='text-base tracking-tight'>
            Two-factor Authentication
          </CardTitle>
          <CardDescription>
            Please enter the authentication code. <br /> We have sent the
            authentication code to your email.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <OtpForm />
        </CardContent>
        <CardFooter className='flex-col gap-2'>
          <p className='px-8 text-center text-sm text-muted-foreground'>
            Haven't received it?{' '}
            <Button
              type='button'
              variant='link'
              className='h-auto px-0 text-sm underline underline-offset-4 hover:text-primary'
              disabled={resendState === 'sending'}
              data-testid='otp-resend'
              onClick={handleResend}
            >
              Resend a new code.
            </Button>
          </p>
          {resendMessage ? (
            <p
              className='px-8 text-center text-sm text-muted-foreground'
              data-testid='otp-resend-feedback'
            >
              {resendMessage}
            </p>
          ) : null}
        </CardFooter>
      </Card>
    </AuthLayout>
  )
}
