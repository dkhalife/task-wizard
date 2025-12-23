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
  private listeners: Map<WSEvent, Set<WebSocketEventListener>> = new Map()
  private pendingRequests: Map<
    string,
    {
      resolve: (value: any) => void
      reject: (reason?: any) => void
      timeoutId: ReturnType<typeof setTimeout>
    }
  > = new Map()
  private dispatch = store.dispatch
  private reconnectTimer?: ReturnType<typeof setTimeout>

  private constructor() {
    this.connect()
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
        const message: {
          action?: string
          data?: unknown
          requestId?: string
          status?: number
        } = JSON.parse(event.data)

        // Request/response messages include requestId + status.
        if (message.requestId && typeof message.status === 'number') {
          const pending = this.pendingRequests.get(message.requestId)
          if (pending) {
            clearTimeout(pending.timeoutId)
            this.pendingRequests.delete(message.requestId)
            pending.resolve(message)
            return
          }
        }

        if (typeof message.action === 'string') {
          this.emit(message.action as WSEvent, message.data as any)
        }
      } catch {
        console.debug('Unexpected WebSocket message type:', event.data)
      }
    }
 
    this.socket.onclose = () => {
      this.socket = null
      this.dispatch(wsDisconnected(null))
      this.rejectAllPendingRequests(new Error('WebSocket connection closed'))
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

    this.rejectAllPendingRequests(new Error('WebSocket disconnected'))

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

  async request<T extends WSAction, TResponse>(
    action: T,
    data?: WSRequest<T>['data'],
    timeoutMs: number = 5000,
  ): Promise<TResponse> {
    if (!this.isConnected()) {
      throw new Error('WebSocket is not connected')
    }

    const requestId = this.newRequestId()
    const request: WSRequest<T> = {
      requestId,
      action,
      ...(data === undefined ? {} : { data }),
    }

    const response = await new Promise<any>((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        this.pendingRequests.delete(requestId)
        reject(new Error('WebSocket request timed out'))
      }, timeoutMs)

      this.pendingRequests.set(requestId, { resolve, reject, timeoutId })
      try {
        this.send(request)
      } catch (e) {
        clearTimeout(timeoutId)
        this.pendingRequests.delete(requestId)
        reject(e)
      }
    })

    const status = response?.status
    const payload = response?.data
    if (typeof status === 'number' && status >= 400) {
      const msg = (payload as any)?.error
      throw new Error(typeof msg === 'string' ? msg : 'WebSocket request failed')
    }

    // Mirror HTTP Request() behavior: treat 204 as empty payload.
    if (status === 204) {
      return {} as TResponse
    }

    return payload as TResponse
  }

  private newRequestId(): string {
    try {
      // Modern browsers
      const c: any = crypto
      if (c && typeof c.randomUUID === 'function') {
        return c.randomUUID()
      }
    } catch {
      // ignore
    }

    return `${Date.now()}-${Math.random().toString(16).slice(2)}`
  }

  private rejectAllPendingRequests(err: Error) {
    const pending = Array.from(this.pendingRequests.values())
    this.pendingRequests.clear()
    pending.forEach(p => {
      clearTimeout(p.timeoutId)
      try {
        p.reject(err)
      } catch {
        // ignore
      }
    })
  }

  private scheduleReconnect() {
    if (this.manualClose) {
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
