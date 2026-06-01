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
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from '@/components/ui/resizable'
import { getUserModels, getUserGroups } from './api'
import { PlaygroundChat } from './components/playground-chat'
import { PlaygroundInput } from './components/playground-input'
import { PlaygroundSidebar } from './components/playground-sidebar'
import { usePlaygroundState, useChatHandler } from './hooks'
import { createUserMessage, createLoadingAssistantMessage } from './lib'
import type { Message as MessageType } from './types'

export function Playground() {
  const { t } = useTranslation()
  const {
    config,
    parameterEnabled,
    messages,
    chatSessions,
    activeChatId,
    models,
    groups,
    updateMessages,
    setModels,
    setGroups,
    updateConfig,
    createChat,
    selectChat,
    deleteChat,
  } = usePlaygroundState()

  const { sendChat, stopGeneration, isGenerating } = useChatHandler({
    config,
    parameterEnabled,
    onMessageUpdate: updateMessages,
  })

  const [editingMessageKey, setEditingMessageKey] = useState<string | null>(
    null
  )
  const [chatSearchQuery, setChatSearchQuery] = useState('')
  const sidebarSearchRef = useRef<HTMLInputElement | null>(null)

  const { data: modelsData, isLoading: isLoadingModels } = useQuery({
    queryKey: ['playground-models'],
    queryFn: async () => {
      try {
        return await getUserModels()
      } catch (error) {
        toast.error(
          error instanceof Error
            ? error.message
            : t('Failed to load playground models')
        )
        return []
      }
    },
  })

  const { data: groupsData } = useQuery({
    queryKey: ['playground-groups'],
    queryFn: async () => {
      try {
        return await getUserGroups()
      } catch (error) {
        toast.error(
          error instanceof Error
            ? error.message
            : t('Failed to load playground groups')
        )
        return []
      }
    },
  })

  useEffect(() => {
    if (!modelsData) return

    setModels(modelsData)

    const isCurrentModelValid = modelsData.some((m) => m.value === config.model)
    if (modelsData.length > 0 && !isCurrentModelValid) {
      updateConfig('model', modelsData[0].value)
    }
  }, [modelsData, config.model, setModels, updateConfig])

  useEffect(() => {
    if (!groupsData) return

    setGroups(groupsData)

    const hasCurrentGroup = groupsData.some((g) => g.value === config.group)
    if (!hasCurrentGroup && groupsData.length > 0) {
      const fallback =
        groupsData.find((g) => g.value === 'default')?.value ??
        groupsData[0].value
      updateConfig('group', fallback)
    }
  }, [groupsData, setGroups, config.group, updateConfig])

  const filteredChats = useMemo(() => {
    const query = chatSearchQuery.trim().toLowerCase()
    if (!query) return chatSessions

    return chatSessions.filter((chat) => {
      if (chat.title.toLowerCase().includes(query)) return true
      return chat.messages.some((message) =>
        message.versions?.some((version) =>
          version.content.toLowerCase().includes(query)
        )
      )
    })
  }, [chatSearchQuery, chatSessions])

  const handleSendMessage = (text: string) => {
    const userMessage = createUserMessage(text)
    const assistantMessage = createLoadingAssistantMessage()

    const newMessages = [...messages, userMessage, assistantMessage]
    updateMessages(newMessages)
    sendChat(newMessages)
  }

  const handleCopyMessage = (message: MessageType) => {
    // eslint-disable-next-line no-console
    console.log('Message copied:', message.key)
  }

  const handleRegenerateMessage = (message: MessageType) => {
    const messageIndex = messages.findIndex((m) => m.key === message.key)
    if (messageIndex === -1) return

    const messagesUpToHere = messages.slice(0, messageIndex)
    const loadingMessage = createLoadingAssistantMessage()
    const newMessages = [...messagesUpToHere, loadingMessage]

    updateMessages(newMessages)
    sendChat(newMessages)
  }

  const handleEditMessage = useCallback((message: MessageType) => {
    setEditingMessageKey(message.key)
  }, [])

  const handleEditOpenChange = useCallback((open: boolean) => {
    if (!open) setEditingMessageKey(null)
  }, [])

  const applyEdit = useCallback(
    (newContent: string, submit: boolean) => {
      if (!editingMessageKey) return
      const index = messages.findIndex((m) => m.key === editingMessageKey)
      if (index === -1) return

      const updated = messages.map((m) =>
        m.key === editingMessageKey
          ? { ...m, versions: [{ ...m.versions[0], content: newContent }] }
          : m
      )

      setEditingMessageKey(null)

      if (!submit || updated[index].from !== 'user') {
        updateMessages(updated)
        return
      }

      const toSubmit = [
        ...updated.slice(0, index + 1),
        createLoadingAssistantMessage(),
      ]
      updateMessages(toSubmit)
      sendChat(toSubmit)
    },
    [editingMessageKey, messages, updateMessages, sendChat]
  )

  const handleDeleteMessage = (message: MessageType) => {
    const newMessages = messages.filter((m) => m.key !== message.key)
    updateMessages(newMessages)
  }

  const handleCreateChat = useCallback(() => {
    if (isGenerating) {
      toast.info(t('Please wait for the current response to finish.'))
      return
    }
    createChat()
    setChatSearchQuery('')
  }, [createChat, isGenerating, t])

  const handleSelectChat = useCallback(
    (chatId: string) => {
      if (isGenerating) {
        toast.info(t('Please wait for the current response to finish.'))
        return
      }
      selectChat(chatId)
    },
    [isGenerating, selectChat, t]
  )

  const handleDeleteChat = useCallback(
    (chatId: string) => {
      if (isGenerating) {
        toast.info(t('Please wait for the current response to finish.'))
        return
      }
      deleteChat(chatId)
    },
    [deleteChat, isGenerating, t]
  )

  const handleSearchClick = useCallback(() => {
    sidebarSearchRef.current?.focus()
    toast.info(t('Search your chat history from the sidebar.'))
  }, [t])

  return (
    <ResizablePanelGroup direction='horizontal' className='h-full min-h-0'>
      <ResizablePanel defaultSize={24} minSize={18} maxSize={32}>
        <div className='h-full min-h-0'>
          <PlaygroundSidebar
            chats={filteredChats}
            activeChatId={activeChatId}
            searchQuery={chatSearchQuery}
            searchInputRef={sidebarSearchRef}
            disabled={isGenerating}
            onSearchQueryChange={setChatSearchQuery}
            onCreateChat={handleCreateChat}
            onSelectChat={handleSelectChat}
            onDeleteChat={handleDeleteChat}
          />
        </div>
      </ResizablePanel>
      <ResizableHandle withHandle />
      <ResizablePanel defaultSize={76} minSize={50}>
        <div className='relative flex size-full flex-col overflow-hidden bg-gradient-to-b from-background to-muted/20'>
          <div className='border-b px-4 py-3 sm:px-6'>
            <div className='mx-auto flex w-full max-w-4xl items-center justify-between gap-3'>
              <div className='min-w-0'>
                <h1 className='truncate text-sm font-semibold sm:text-base'>
                  {
                    chatSessions.find((chat) => chat.id === activeChatId)?.title ||
                    t('New Chat')
                  }
                </h1>
                <p className='text-muted-foreground text-xs sm:text-sm'>
                  {t('Persistent chat playground with searchable conversations.')}
                </p>
              </div>
            </div>
          </div>

          <div className='flex flex-1 flex-col overflow-hidden'>
            <PlaygroundChat
              messages={messages}
              onCopyMessage={handleCopyMessage}
              onRegenerateMessage={handleRegenerateMessage}
              onEditMessage={handleEditMessage}
              onDeleteMessage={handleDeleteMessage}
              isGenerating={isGenerating}
              editingKey={editingMessageKey}
              onCancelEdit={handleEditOpenChange}
              onSaveEdit={(newContent) => applyEdit(newContent, false)}
              onSaveEditAndSubmit={(newContent) => applyEdit(newContent, true)}
            />
          </div>

          <div className='mx-auto w-full max-w-4xl px-3 pb-3 sm:px-4 sm:pb-4'>
            <PlaygroundInput
              disabled={isGenerating}
              groups={groups}
              groupValue={config.group}
              isGenerating={isGenerating}
              isModelLoading={isLoadingModels}
              modelValue={config.model}
              models={models}
              onGroupChange={(value) => updateConfig('group', value)}
              onModelChange={(value) => updateConfig('model', value)}
              onSearchClick={handleSearchClick}
              onStop={stopGeneration}
              onSubmit={handleSendMessage}
              showSuggestions={messages.length === 0}
            />
          </div>
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>
  )
}
