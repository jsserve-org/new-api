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
import type { RefObject } from 'react'
import { MessageSquarePlus, Search, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { ChatSession } from '../types'

type PlaygroundSidebarProps = {
  chats: ChatSession[]
  activeChatId: string
  searchQuery: string
  searchInputRef?: RefObject<HTMLInputElement | null>
  disabled?: boolean
  onSearchQueryChange: (value: string) => void
  onCreateChat: () => void
  onSelectChat: (chatId: string) => void
  onDeleteChat: (chatId: string) => void
}

function formatTimestamp(timestamp: number) {
  return new Date(timestamp).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

export function PlaygroundSidebar(props: PlaygroundSidebarProps) {
  const { t } = useTranslation()

  return (
    <aside className='bg-background/95 flex h-full flex-col border-r backdrop-blur'>
      <div className='space-y-3 border-b p-4'>
        <div className='flex items-center justify-between gap-2'>
          <div>
            <h2 className='text-sm font-semibold'>{t('Chats')}</h2>
            <p className='text-muted-foreground text-xs'>
              {t('Create, search, and switch between conversations.')}
            </p>
          </div>
          <Button
            size='sm'
            className='gap-1.5'
            onClick={props.onCreateChat}
            disabled={props.disabled}
          >
            <MessageSquarePlus className='size-4' />
            {t('New Chat')}
          </Button>
        </div>

        <div className='relative'>
          <Search className='text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2' />
          <Input
            ref={props.searchInputRef}
            value={props.searchQuery}
            disabled={props.disabled}
            onChange={(event) => props.onSearchQueryChange(event.target.value)}
            placeholder={t('Search chats...')}
            className='h-9 pl-9'
          />
        </div>
      </div>

      <div className='flex-1 space-y-1 overflow-y-auto p-2'>
        {props.chats.length === 0 ? (
          <div className='text-muted-foreground px-3 py-8 text-center text-sm'>
            {t('No chats found')}
          </div>
        ) : (
          props.chats.map((chat) => {
            const preview =
              chat.messages[chat.messages.length - 1]?.versions?.[0]?.content ||
              t('Start a new conversation')

            return (
              <div
                key={chat.id}
                role='button'
                tabIndex={0}
                onClick={() => {
                  if (!props.disabled) props.onSelectChat(chat.id)
                }}
                onKeyDown={(event) => {
                  if (event.key === 'Enter' || event.key === ' ') {
                    event.preventDefault()
                    if (!props.disabled) props.onSelectChat(chat.id)
                  }
                }}
                className={cn(
                  'group hover:bg-muted/70 flex w-full cursor-pointer flex-col rounded-xl border px-3 py-3 text-left transition-colors',
                  props.activeChatId === chat.id
                    ? 'border-primary/40 bg-primary/5'
                    : 'border-transparent'
                )}
              >
                <div className='flex items-start gap-2'>
                  <div className='min-w-0 flex-1'>
                    <div className='truncate text-sm font-medium'>
                      {chat.title}
                    </div>
                    <div className='text-muted-foreground mt-1 line-clamp-2 text-xs leading-relaxed'>
                      {preview}
                    </div>
                  </div>
                  <Button
                    size='icon'
                    variant='ghost'
                    className='text-muted-foreground hover:text-destructive size-7 opacity-0 transition-opacity group-hover:opacity-100'
                    onClick={(event) => {
                      event.stopPropagation()
                      if (!props.disabled) props.onDeleteChat(chat.id)
                    }}
                    disabled={props.disabled}
                    title={t('Delete')}
                  >
                    <Trash2 className='size-4' />
                  </Button>
                </div>
                <div className='text-muted-foreground mt-2 text-[11px]'>
                  {formatTimestamp(chat.updatedAt)}
                </div>
              </div>
            )
          })
        )}
      </div>
    </aside>
  )
}
