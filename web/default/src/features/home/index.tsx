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
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { Markdown } from '@/components/ui/markdown'
import { PublicLayout } from '@/components/layout'
import { useHomePageContent } from './hooks'
import { Link } from '@tanstack/react-router'

export function Home() {
  const { t } = useTranslation()
  const { auth } = useAuthStore()
  const isAuthenticated = !!auth.user
  const { content, isLoaded, isUrl } = useHomePageContent()

  if (!isLoaded) {
    return (
      <PublicLayout showMainContainer={false}>
        <main className='flex min-h-screen items-center justify-center'>
          <div className='text-muted-foreground'>{t('Loading...')}</div>
        </main>
      </PublicLayout>
    )
  }

  if (content) {
    return (
      <PublicLayout showMainContainer={false}>
        <main className='overflow-x-hidden'>
          {isUrl ? (
            <iframe
              src={content}
              className='h-screen w-full border-none'
              title={t('Custom Home Page')}
            />
          ) : (
            <div className='container mx-auto py-8'>
              <Markdown className='custom-home-content'>{content}</Markdown>
            </div>
          )}
        </main>
      </PublicLayout>
    )
  }

  return (
    <PublicLayout showMainContainer={false}>
      <main className='flex min-h-screen flex-col items-center justify-center px-4'>
        <div className='max-w-2xl text-center'>
          <h1 className='mb-4 text-4xl font-bold tracking-tight'>
            {t('AI API Gateway')}
          </h1>
          <p className='text-muted-foreground mb-8 text-lg leading-relaxed'>
            {t(
              'A unified API gateway that aggregates 40+ AI providers behind a single endpoint. Manage users, billing, rate limiting, and more.'
            )}
          </p>
          <div className='flex items-center justify-center gap-4'>
            {isAuthenticated ? (
              <Link
                to='/playground'
                className='bg-primary text-primary-foreground hover:bg-primary/90 inline-flex h-10 items-center justify-center rounded-md px-6 text-sm font-medium transition-colors'
              >
                {t('Go to Playground')}
              </Link>
            ) : (
              <>
                <Link
                  to='/login'
                  className='bg-primary text-primary-foreground hover:bg-primary/90 inline-flex h-10 items-center justify-center rounded-md px-6 text-sm font-medium transition-colors'
                >
                  {t('Sign In')}
                </Link>
                <Link
                  to='/register'
                  className='border-input hover:bg-accent hover:text-accent-foreground inline-flex h-10 items-center justify-center rounded-md border px-6 text-sm font-medium transition-colors'
                >
                  {t('Sign Up')}
                </Link>
              </>
            )}
          </div>
        </div>
      </main>
    </PublicLayout>
  )
}