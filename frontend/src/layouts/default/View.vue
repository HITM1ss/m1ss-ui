<template>
  <v-main
    id="panel-main"
    tabindex="-1"
  >
    <div class="panel-content">
      <router-view />
    </div>
  </v-main>
</template>

<style>
.panel-content { max-width: 1480px; margin-inline: auto; padding: 32px 36px 48px; }
@media (max-width: 960px) { .panel-content { padding: 20px; } }
@media (max-width: 600px) { .panel-content { padding: 16px 12px 24px; } }
</style>
