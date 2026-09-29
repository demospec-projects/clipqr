<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import Icone from './Icone.vue'

defineProps<{ apercu: string; image: string; erreur: string }>()
const emit = defineEmits<{ fermer: [] }>()

function surTouche(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('fermer')
}
onMounted(() => window.addEventListener('keydown', surTouche))
onBeforeUnmount(() => window.removeEventListener('keydown', surTouche))
</script>

<template>
  <div class="voile" @click.self="emit('fermer')">
    <section class="feuille" role="dialog" aria-label="QR code">
      <header>
        <strong>Scanner avec le téléphone</strong>
        <button class="fermer" type="button" title="Fermer (Échap)" @click="emit('fermer')"><Icone nom="fermer" /></button>
      </header>
      <div v-if="erreur" class="erreur">
        <Icone nom="alerte" :taille="28" />
        <p>{{ erreur }}</p>
      </div>
      <template v-else>
        <div class="cadre">
          <img v-if="image" :src="image" alt="QR code du texte" />
          <div v-else class="attente" />
        </div>
        <p class="apercu">{{ apercu }}</p>
        <p class="aide">Le texte est dans le code : aucune connexion requise.</p>
      </template>
    </section>
  </div>
</template>

<style scoped>
.voile {
  position: absolute;
  inset: 0;
  z-index: 10;
  display: grid;
  place-items: center;
  padding: 20px;
  background: var(--voile);
  backdrop-filter: blur(6px);
}
.feuille {
  width: 100%;
  border-radius: 18px;
  background: var(--fond-eleve);
  border: 1px solid var(--bordure);
  box-shadow: 0 24px 48px -12px rgb(0 0 0 / 0.45);
  padding: 14px 16px 16px;
  animation: monte 0.18s ease-out;
}
header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  font-size: 14px;
}
.fermer {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border-radius: 8px;
  color: var(--discret);
}
.fermer:hover {
  background: var(--survol);
  color: var(--texte);
}
.cadre {
  aspect-ratio: 1;
  border-radius: 14px;
  background: #fff;
  padding: 10px;
  box-shadow: inset 0 0 0 1px rgb(0 0 0 / 0.06);
}
.cadre img {
  width: 100%;
  height: 100%;
  image-rendering: pixelated;
  display: block;
}
.attente {
  width: 100%;
  height: 100%;
  border-radius: 8px;
  background: linear-gradient(90deg, #eef2ff, #f5f3ff, #eef2ff);
  background-size: 200% 100%;
  animation: reflet 1s linear infinite;
}
.apercu {
  margin: 12px 0 4px;
  font-size: 12.5px;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}
.aide {
  margin: 0;
  font-size: 11.5px;
  color: var(--discret);
}
.erreur {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 8px 8px;
  text-align: center;
  color: var(--attention);
}
.erreur p {
  margin: 0;
  color: var(--texte);
  font-size: 13px;
}
@keyframes monte {
  from {
    opacity: 0;
    transform: translateY(8px) scale(0.98);
  }
}
@keyframes reflet {
  to {
    background-position: -200% 0;
  }
}
</style>
