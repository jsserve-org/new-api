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
import { useState, useCallback, useMemo, useEffect, useRef } from 'react'
import { nanoid } from 'nanoid'
import { DEFAULT_CONFIG, DEFAULT_PARAMETER_ENABLED } from '../constants'
import {
  loadConfig,
  saveConfig,
  loadParameterEnabled,
  saveParameterEnabled,
  saveMessages,
  loadChats,
  saveChats,
  loadActiveChatId,
  saveActiveChatId,
} from '../lib'
import type {
  Message,
  PlaygroundConfig,
  ParameterEnabled,
  ModelOption,
  GroupOption,
  ChatSession,
} from '../types'

const DEFAULT_CHAT_TITLE = 'New Chat'

function createEmptyChat(): ChatSession {
  const now = Date.now()
  return {
    id: nanoid(),
    title: DEFAULT_CHAT_TITLE,
    createdAt: now,
    updatedAt: now,
    messages: [],
  }
}

function getChatTitle(messages: Message[]): string {
  const firstUserMessage = messages.find((message) => message.from === 'user')
  const rawTitle = firstUserMessage?.versions?.[0]?.content?.trim()
  if (!rawTitle) return DEFAULT_CHAT_TITLE
  return rawTitle.replace(/\s+/g, ' ').slice(0, 48)
}

function loadInitialChatState(): {
  chats: ChatSession[]
  activeChatId: string
} {
  const chats = loadChats()
  const savedActiveChatId = loadActiveChatId()

  if (chats.length > 0) {
    return {
      chats,
      activeChatId:
        savedActiveChatId && chats.some((chat) => chat.id === savedActiveChatId)
          ? savedActiveChatId
          : chats[0].id,
    }
  }

  const fallbackChat = createEmptyChat()
  saveChats([fallbackChat])
  saveActiveChatId(fallbackChat.id)
  return {
    chats: [fallbackChat],
    activeChatId: fallbackChat.id,
  }
}

/**
 * Main state management hook for playground
 */
export function usePlaygroundState() {
  const [config, setConfig] = useState<PlaygroundConfig>(() => {
    const savedConfig = loadConfig()
    return { ...DEFAULT_CONFIG, ...savedConfig }
  })

  const [parameterEnabled, setParameterEnabled] = useState<ParameterEnabled>(
    () => {
      const saved = loadParameterEnabled()
      return { ...DEFAULT_PARAMETER_ENABLED, ...saved }
    }
  )

  const [chatSessions, setChatSessions] = useState<ChatSession[]>(
    () => loadInitialChatState().chats
  )

  const [activeChatId, setActiveChatId] = useState<string>(
    () => loadInitialChatState().activeChatId
  )

  const [models, setModels] = useState<ModelOption[]>([])
  const [groups, setGroups] = useState<GroupOption[]>([])
  const activeChatIdRef = useRef(activeChatId)

  useEffect(() => {
    activeChatIdRef.current = activeChatId
    saveActiveChatId(activeChatId)
  }, [activeChatId])

  useEffect(() => {
    saveChats(chatSessions)
    const selectedChat =
      chatSessions.find((chat) => chat.id === activeChatId) ?? chatSessions[0]
    saveMessages(selectedChat?.messages ?? [])
  }, [activeChatId, chatSessions])

  const activeChat = useMemo(() => {
    return (
      chatSessions.find((chat) => chat.id === activeChatId) ?? chatSessions[0]
    )
  }, [chatSessions, activeChatId])

  const messages = activeChat?.messages ?? []

  const updateConfig = useCallback(
    <K extends keyof PlaygroundConfig>(key: K, value: PlaygroundConfig[K]) => {
      setConfig((prev) => {
        const updated = { ...prev, [key]: value }
        saveConfig(updated)
        return updated
      })
    },
    []
  )

  const updateParameterEnabled = useCallback(
    (key: keyof ParameterEnabled, value: boolean) => {
      setParameterEnabled((prev) => {
        const updated = { ...prev, [key]: value }
        saveParameterEnabled(updated)
        return updated
      })
    },
    []
  )

  const updateMessages = useCallback(
    (updater: Message[] | ((prev: Message[]) => Message[])) => {
      setChatSessions((prev) => {
        const nextChats = prev.map((chat) => {
          if (chat.id !== activeChatIdRef.current) return chat
          const nextMessages =
            typeof updater === 'function' ? updater(chat.messages) : updater
          return {
            ...chat,
            messages: nextMessages,
            updatedAt: Date.now(),
            title: getChatTitle(nextMessages),
          }
        })

        return nextChats.sort((a, b) => b.updatedAt - a.updatedAt)
      })
    },
    []
  )

  const createChat = useCallback(() => {
    const newChat = createEmptyChat()
    setChatSessions((prev) => [newChat, ...prev])
    setActiveChatId(newChat.id)
  }, [])

  const selectChat = useCallback(
    (chatId: string) => {
      if (!chatSessions.some((chat) => chat.id === chatId)) return
      setActiveChatId(chatId)
    },
    [chatSessions]
  )

  const deleteChat = useCallback(
    (chatId: string) => {
      const remainingChats = chatSessions.filter((chat) => chat.id !== chatId)
      if (remainingChats.length === 0) {
        const fallbackChat = createEmptyChat()
        setChatSessions([fallbackChat])
        setActiveChatId(fallbackChat.id)
        return
      }

      setChatSessions(remainingChats)
      if (activeChatId === chatId) {
        setActiveChatId(remainingChats[0].id)
      }
    },
    [activeChatId, chatSessions]
  )

  const clearMessages = useCallback(() => {
    updateMessages([])
  }, [updateMessages])

  const resetConfig = useCallback(() => {
    setConfig(DEFAULT_CONFIG)
    setParameterEnabled(DEFAULT_PARAMETER_ENABLED)
    saveConfig(DEFAULT_CONFIG)
    saveParameterEnabled(DEFAULT_PARAMETER_ENABLED)
  }, [])

  return {
    config,
    parameterEnabled,
    messages,
    chatSessions,
    activeChatId,
    activeChat,
    models,
    groups,
    setModels,
    setGroups,
    updateConfig,
    updateParameterEnabled,
    updateMessages,
    clearMessages,
    resetConfig,
    createChat,
    selectChat,
    deleteChat,
  }
}
