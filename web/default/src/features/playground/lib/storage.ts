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
import { nanoid } from 'nanoid'
import { STORAGE_KEYS } from '../constants'
import type {
  PlaygroundConfig,
  ParameterEnabled,
  Message,
  ChatSession,
} from '../types'
import { sanitizeMessagesOnLoad } from './message-utils'

function sanitizeChatSession(session: ChatSession): ChatSession {
  return {
    ...session,
    title: session.title || 'New Chat',
    createdAt: Number(session.createdAt) || Date.now(),
    updatedAt: Number(session.updatedAt) || Date.now(),
    messages: sanitizeMessagesOnLoad(session.messages || []),
  }
}

function createMigratedChat(messages: Message[]): ChatSession {
  const now = Date.now()
  return {
    id: nanoid(),
    title: 'New Chat',
    createdAt: now,
    updatedAt: now,
    messages: sanitizeMessagesOnLoad(messages),
  }
}

/**
 * Load playground config from localStorage
 */
export function loadConfig(): Partial<PlaygroundConfig> {
  try {
    const saved = localStorage.getItem(STORAGE_KEYS.CONFIG)
    if (saved) {
      return JSON.parse(saved)
    }
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to load config:', error)
  }
  return {}
}

/**
 * Save playground config to localStorage
 */
export function saveConfig(config: Partial<PlaygroundConfig>): void {
  try {
    localStorage.setItem(STORAGE_KEYS.CONFIG, JSON.stringify(config))
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to save config:', error)
  }
}

/**
 * Load parameter enabled state from localStorage
 */
export function loadParameterEnabled(): Partial<ParameterEnabled> {
  try {
    const saved = localStorage.getItem(STORAGE_KEYS.PARAMETER_ENABLED)
    if (saved) {
      return JSON.parse(saved)
    }
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to load parameter enabled:', error)
  }
  return {}
}

/**
 * Save parameter enabled state to localStorage
 */
export function saveParameterEnabled(
  parameterEnabled: Partial<ParameterEnabled>
): void {
  try {
    localStorage.setItem(
      STORAGE_KEYS.PARAMETER_ENABLED,
      JSON.stringify(parameterEnabled)
    )
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to save parameter enabled:', error)
  }
}

/**
 * Load messages from localStorage
 */
export function loadMessages(): Message[] | null {
  try {
    const saved = localStorage.getItem(STORAGE_KEYS.MESSAGES)
    if (saved) {
      const parsed: unknown = JSON.parse(saved)
      if (!Array.isArray(parsed)) {
        return null
      }
      const sanitized = sanitizeMessagesOnLoad(parsed as Message[])
      if (sanitized !== parsed) {
        saveMessages(sanitized)
      }
      return sanitized
    }
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to load messages:', error)
  }
  return null
}

/**
 * Save messages to localStorage
 */
export function saveMessages(messages: Message[]): void {
  try {
    localStorage.setItem(STORAGE_KEYS.MESSAGES, JSON.stringify(messages))
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to save messages:', error)
  }
}

export function loadChats(): ChatSession[] {
  try {
    const saved = localStorage.getItem(STORAGE_KEYS.CHATS)
    if (saved) {
      const parsed: unknown = JSON.parse(saved)
      if (Array.isArray(parsed)) {
        const chats = parsed.map((chat) => sanitizeChatSession(chat as ChatSession))
        saveChats(chats)
        return chats
      }
    }

    const legacyMessages = loadMessages()
    if (legacyMessages?.length) {
      const migrated = [createMigratedChat(legacyMessages)]
      saveChats(migrated)
      saveActiveChatId(migrated[0].id)
      return migrated
    }
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to load chats:', error)
  }

  return []
}

export function saveChats(chats: ChatSession[]): void {
  try {
    localStorage.setItem(STORAGE_KEYS.CHATS, JSON.stringify(chats))
    const activeChat = chats.find((chat) => chat.messages.length > 0) ?? chats[0]
    if (activeChat) {
      saveMessages(activeChat.messages)
    }
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to save chats:', error)
  }
}

export function loadActiveChatId(): string | null {
  try {
    return localStorage.getItem(STORAGE_KEYS.ACTIVE_CHAT_ID)
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to load active chat id:', error)
  }
  return null
}

export function saveActiveChatId(chatId: string): void {
  try {
    localStorage.setItem(STORAGE_KEYS.ACTIVE_CHAT_ID, chatId)
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to save active chat id:', error)
  }
}

/**
 * Clear all playground data
 */
export function clearPlaygroundData(): void {
  try {
    localStorage.removeItem(STORAGE_KEYS.CONFIG)
    localStorage.removeItem(STORAGE_KEYS.MESSAGES)
    localStorage.removeItem(STORAGE_KEYS.CHATS)
    localStorage.removeItem(STORAGE_KEYS.ACTIVE_CHAT_ID)
    localStorage.removeItem(STORAGE_KEYS.PARAMETER_ENABLED)
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to clear playground data:', error)
  }
}
