/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { SettingsSection } from '../components/settings-section'
import { confirmPaymentCompliance } from '../api'

type PaymentBalanceOnlySectionProps = {
  complianceConfirmed: boolean
}

export function PaymentBalanceOnlySection(
  props: PaymentBalanceOnlySectionProps
) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const confirmComplianceMutation = useMutation({
    mutationFn: confirmPaymentCompliance,
    onSuccess: (data) => {
      if (data.success) {
        toast.success(t('Compliance confirmed successfully'))
        queryClient.invalidateQueries({ queryKey: ['system-options'] })
      } else {
        toast.error(data.message || t('Failed to confirm compliance'))
      }
    },
    onError: (error: Error) => {
      toast.error(error.message || t('Failed to confirm compliance'))
    },
  })

  return (
    <SettingsSection title={t('Balance Management')}>
      <Alert>
        <AlertTitle>{t('Settings-based payment methods removed')}</AlertTitle>
        <AlertDescription>
          {t(
            'Online top-up gateway configuration has been removed from Settings. Add balance manually from the Users tab instead.'
          )}
        </AlertDescription>
      </Alert>

      <div className='flex items-start gap-3 rounded-lg border p-4'>
        <Checkbox
          id='payment-compliance-confirmed'
          checked={props.complianceConfirmed}
          disabled={props.complianceConfirmed || confirmComplianceMutation.isPending}
          onCheckedChange={(checked) => {
            if (checked && !props.complianceConfirmed) {
              confirmComplianceMutation.mutate()
            }
          }}
        />
        <div className='space-y-1'>
          <Label
            htmlFor='payment-compliance-confirmed'
            className='cursor-pointer'
          >
            {t(
              'I have read and understood the compliance reminder for operating paid AI services.'
            )}
          </Label>
          <p className='text-muted-foreground text-sm'>
            {t(
              'This checkbox replaces the previous compliance confirmation dialog.'
            )}
          </p>
        </div>
      </div>
    </SettingsSection>
  )
}
