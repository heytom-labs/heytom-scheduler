<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { wsService } from '@/api/websocket'

onMounted(() => {
  // Initialize WebSocket connection when app starts
  wsService.connect()
})

onUnmounted(() => {
  // Clean up WebSocket connection when app unmounts
  wsService.disconnect()
})
</script>

<style>
#app {
  width: 100%;
  height: 100vh;
  margin: 0;
  padding: 0;
}

body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}
</style>
