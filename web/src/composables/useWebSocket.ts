import { onMounted, onUnmounted } from 'vue'
import { wsService, WebSocketMessage } from '@/api/websocket'

export function useWebSocket() {
  onMounted(() => {
    // Connect to WebSocket when component mounts
    wsService.connect()
  })

  onUnmounted(() => {
    // Disconnect when component unmounts
    wsService.disconnect()
  })

  const subscribe = (event: string, callback: (message: WebSocketMessage) => void) => {
    wsService.on(event, callback)
    
    // Return unsubscribe function
    return () => {
      wsService.off(event, callback)
    }
  }

  return {
    connected: wsService.connected,
    subscribe,
    send: wsService.send.bind(wsService)
  }
}
