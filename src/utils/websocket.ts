import {
  WSEvent,
  WSAction,
  WSEventPayloads,
  WSRequest,
} from '@/models/websocket'
import { store } from '@/store/store'
import { wsConnecting, wsConnected, wsDisconnected } from '@/store/wsSlice'

const API_URL = import.meta.env.VITE_APP_API_URL

type WebSocketEventListener = (data: WSEventPayloads[WSEvent]) => void

export class WebSocketManager {
  private static instance: WebSocketManager
  private socket: WebSocket | null = null
  private retryCount = 0
  private manualClose = false
  private enabled: boolean = store.getState().featureFlags.useWebsockets
  private listeners: Map<WSEvent, Set<WebSocketEventListener>> = new Map()
  private dispatch = store.dispatch
  private reconnectTimer?: ReturnType<typeof setTimeout>

  private constructor() {
    if (this.enabled) {
      this.connect()
    }

    store.subscribe(() => {
      const newState = store.getState()
      const newEnabledState = newState.featureFlags.useWebsockets

      // If websockets were disabled, disconnect
      if (this.enabled && !newEnabledState) {
        this.enabled = newEnabledState
        this.disconnect()
      }
      // If websockets were enabled, try to connect
      else if (!this.enabled && newEnabledState) {
        this.enabled = newEnabledState
        this.connect()
      }
    })
  }

  static getInstance(): WebSocketManager {
    if (!WebSocketManager.instance) {
      WebSocketManager.instance = new WebSocketManager()
    }
    return WebSocketManager.instance
  }

  connect() {
    const token = localStorage.getItem('ca_token')
    if (!token) {
      this.dispatch(wsDisconnected("User is not signed in"))
      return
    }

    if (
      this.socket &&
      (this.socket.readyState === WebSocket.OPEN ||
        this.socket.readyState === WebSocket.CONNECTING)
    ) {
      return
    }

    this.dispatch(wsConnecting())
    this.manualClose = false

    const wsProtocol = API_URL.startsWith('https') ? 'wss' : 'ws'
    const baseUrl = API_URL.replace(/^https?/, wsProtocol)
    const url = `${baseUrl}/ws`

    try {
      this.socket = new WebSocket(url, [wsProtocol, token])
    } catch (err: any) {
      this.dispatch(wsDisconnected(err.message || 'Failed to connect'))
      this.scheduleReconnect()
      return
    }

    this.socket.onopen = () => {
      this.retryCount = 0
      this.dispatch(wsConnected())
    }

    this.socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data)
        this.emit(message.action, message.data)
      } catch {
        console.debug('Unexpected WebSocket message type:', event.data)
      }
    }
 
    this.socket.onclose = () => {
      this.socket = null
      this.dispatch(wsDisconnected(null))
      this.scheduleReconnect()
    }

    this.socket.onerror = (event) => {
      console.debug('WebSocket error:', event)
      this.dispatch(wsDisconnected('error'))
    }
  }

  on<T extends WSEvent>(action: T, handler: (data: WSEventPayloads[T]) => void) {
    if (!this.listeners.has(action)) {
      this.listeners.set(action, new Set())
    }

    this.listeners
      .get(action)!
      .add(handler as WebSocketEventListener)
  }

  off<T extends WSEvent>(action: T, handler: (data: WSEventPayloads[T]) => void) {
    const handlers = this.listeners.get(action)
    if (!handlers) {
      return
    }

    handlers.delete(handler as WebSocketEventListener)
    if (handlers.size === 0) {
      this.listeners.delete(action)
    }
  }

  async waitFor<T extends WSEvent>(
    event: T,
    condition: (data: WSEventPayloads[T]) => boolean,
  ): Promise<WSEventPayloads[T]> {
    return new Promise((resolve, reject) => {
      const handler = (data: WSEventPayloads[T]) => {
        let result = false
        try {
          result = condition(data)
        } catch (e) {
          console.debug('WebSocket waitFor condition error', e)
        }

        if (result) {
          clearTimeout(timeoutId)
          this.off(event, handler)
          resolve(data)
        }
      }

      const timeoutId = setTimeout(() => {
        this.off(event, handler)
        reject(new Error('Timed out waiting for condition'))
      }, 5000)

      this.on(event, handler)
    })
  }

  private emit<T extends WSEvent>(event: T, data: WSEventPayloads[T]) {
    const handlers = this.listeners.get(event)
    if (!handlers) {
      return
    }

    handlers.forEach(h => {
      try {
        h(data)
      } catch (e) {
        console.debug('WebSocket listener error', e)
      }
    })
  }

  disconnect() {
    this.manualClose = true
    if (this.socket) {
      this.socket.close()
      this.socket = null
      this.dispatch(wsDisconnected(null))
    }

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = undefined
    }

    this.retryCount = 0
  }

  isConnected(): boolean {
    return this.socket !== null && this.socket.readyState === WebSocket.OPEN
  }

  send<T extends WSAction>(request: WSRequest<T>): void {
    if (!this.isConnected()) {
      throw new Error('WebSocket is not connected')
    }

    this.socket!.send(JSON.stringify(request))
  }

  private scheduleReconnect() {
    if (this.manualClose || !this.enabled) {
      return
    }

    const delay = Math.min(30000, Math.pow(2, this.retryCount) * 1000)
    this.retryCount += 1
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }
    this.reconnectTimer = setTimeout(() => this.connect(), delay)
  }
}

export default WebSocketManager
