<template>
  <v-main>
    <div class="panel-content">
      <router-view />
    </div>
  </v-main>
</template>

<style>
.panel-content { max-width: 1680px; margin-inline: auto; padding: 28px 32px 36px; }
@media (max-width: 960px) { .panel-content { padding: 20px; } }
@media (max-width: 600px) { .panel-content { padding: 16px 12px 24px; } }
</style>
