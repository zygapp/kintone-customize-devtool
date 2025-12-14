<script setup lang="ts">
import { ref } from 'vue'

const count = ref(0)
</script>

<template>
  <div class="kcdev-app">
    <h1>kintone Customize</h1>
    <button @click="count++">
      count is {{ count }}
    </button>
  </div>
</template>
