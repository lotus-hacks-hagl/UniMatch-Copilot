import { ref, onMounted, onUnmounted } from 'vue'

export function usePolling(callback, intervalMs = 5000) {
  let intervalId = null

  const start = () => {
    if (!intervalId) {
      intervalId = setInterval(callback, intervalMs)
    }
  }

  const stop = () => {
    if (intervalId) {
      clearInterval(intervalId)
      intervalId = null
    }
  }

  onMounted(start)
  onUnmounted(stop)

  return { start, stop }
}
