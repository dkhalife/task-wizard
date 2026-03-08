import WebSocketManager from '@/utils/websocket'

export async function transport<T>(opts: {
  http: () => Promise<T>
  ws: (mgr: WebSocketManager) => Promise<T>
}): Promise<T> {
  const mgr = WebSocketManager.getInstance()
  if (mgr.isConnected()) {
    return opts.ws(mgr)
  }

  return opts.http()
}
