/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.
*/
import { useEffect, useState } from 'react'
import { Check, Copy, ExternalLink, Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  completeCopilotDeviceAuth,
  startCopilotDeviceAuth,
} from '../../api'

type CopilotDeviceAuthDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onKeyGenerated: (key: string) => void
}

export function CopilotDeviceAuthDialog({
  open,
  onOpenChange,
  onKeyGenerated,
}: CopilotDeviceAuthDialogProps) {
  const { t } = useTranslation()
  const { copiedText, copyToClipboard } = useCopyToClipboard({ notify: false })
  const [state, setState] = useState({
    deviceCode: '',
    userCode: '',
    verificationUri: '',
    verificationUriComplete: '',
    isStarting: false,
    isCompleting: false,
  })

  useEffect(() => {
    if (!open) {
      setState({
        deviceCode: '',
        userCode: '',
        verificationUri: '',
        verificationUriComplete: '',
        isStarting: false,
        isCompleting: false,
      })
    }
  }, [open])

  const handleStart = async () => {
    setState((prev) => ({ ...prev, isStarting: true }))
    try {
      const res = await startCopilotDeviceAuth()
      if (!res.success) throw new Error(res.message || 'Failed to start device authorization')
      const data = res.data
      if (!data?.device_code || !data.user_code || !data.verification_uri) {
        throw new Error('Missing device authorization fields')
      }
      setState((prev) => ({
        ...prev,
        deviceCode: data.device_code || '',
        userCode: data.user_code || '',
        verificationUri: data.verification_uri || '',
        verificationUriComplete: data.verification_uri_complete || '',
      }))
      window.open(
        data.verification_uri_complete || data.verification_uri,
        '_blank',
        'noopener,noreferrer'
      )
      toast.success(t('Opened GitHub device authorization page'))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('Device authorization failed'))
    } finally {
      setState((prev) => ({ ...prev, isStarting: false }))
    }
  }

  const handleComplete = async () => {
    if (!state.deviceCode) return
    setState((prev) => ({ ...prev, isCompleting: true }))
    try {
      const res = await completeCopilotDeviceAuth(state.deviceCode)
      if (!res.success) throw new Error(res.message || 'Authorization pending or failed')
      const key = res.data?.key || ''
      if (!key) throw new Error('Missing GitHub OAuth token')
      onKeyGenerated(key)
      toast.success(t('GitHub Copilot credential generated'))
      onOpenChange(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('Authorization pending or failed'))
    } finally {
      setState((prev) => ({ ...prev, isCompleting: false }))
    }
  }

  const authUrl = state.verificationUriComplete || state.verificationUri

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>{t('GitHub Copilot Device Authorization')}</DialogTitle>
          <DialogDescription>
            {t('Authorize with GitHub device-code flow to generate a Copilot channel token.')}
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4'>
          <Alert>
            <AlertDescription>
              {t('Click Start, sign in on GitHub, enter the device code, then click Complete authorization.')}
            </AlertDescription>
          </Alert>

          <div className='flex flex-wrap gap-2'>
            <Button onClick={handleStart} disabled={state.isStarting}>
              {state.isStarting ? <Loader2 className='mr-2 h-4 w-4 animate-spin' /> : <ExternalLink className='mr-2 h-4 w-4' />}
              {t('Start GitHub authorization')}
            </Button>
            <Button
              type='button'
              variant='outline'
              disabled={!authUrl}
              onClick={() => authUrl && window.open(authUrl, '_blank', 'noopener,noreferrer')}
            >
              <ExternalLink className='mr-2 h-4 w-4' />
              {t('Open GitHub')}
            </Button>
          </div>

          {state.userCode && (
            <div className='rounded-md border p-4 text-center'>
              <div className='text-muted-foreground mb-2 text-xs'>{t('Device code')}</div>
              <div className='font-mono text-2xl font-semibold tracking-widest'>{state.userCode}</div>
              <Button
                type='button'
                variant='ghost'
                size='sm'
                className='mt-2'
                onClick={() => copyToClipboard(state.userCode)}
              >
                {copiedText === state.userCode ? <Check className='mr-2 h-4 w-4 text-green-600' /> : <Copy className='mr-2 h-4 w-4' />}
                {t('Copy code')}
              </Button>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button type='button' variant='outline' onClick={() => onOpenChange(false)} disabled={state.isStarting || state.isCompleting}>
            {t('Cancel')}
          </Button>
          <Button onClick={handleComplete} disabled={!state.deviceCode || state.isCompleting}>
            {state.isCompleting && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
            {state.isCompleting ? t('Completing...') : t('Complete authorization')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
