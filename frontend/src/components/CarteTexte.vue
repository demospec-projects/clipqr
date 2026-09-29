<script setup lang="ts">
import { computed } from 'vue'
import type { EntryView } from '../../bindings/github.com/demospec-projects/clipqr/models'
import { quand } from '../temps'
import Icone from './Icone.vue'

const props = defineProps<{
  entree: EntryView
  // Le texte est dans le presse-papiers : un Ctrl+V ailleurs le colle.
  actuel: boolean
  // Il vient d'être copié depuis ClipQR (confirmation brève).
  copie: boolean
  maintenant: number
  raccourci: string
}>()
const emit = defineEmits<{ copier: []; epingler: []; qr: []; ouvrir: [] }>()

const libelles: Record<string, string> = { link: 'Lien', email: 'Courriel', phone: 'Téléphone', text: 'Texte' }

const apercu = computed(() => props.entree.text.trim() + (props.entree.truncated ? '…' : ''))
const ouvrable = computed(() => props.entree.kind === 'link' || props.entree.kind === 'email')
const libelle = computed(() => {
  const lignes = props.entree.text.trim().split('\n').length
  return props.entree.kind === 'text' && lignes > 1 ? `${lignes} lignes` : libelles[props.entree.kind] ?? 'Texte'
})
const moment = computed(() => quand(props.entree.copiedAt, props.maintenant))
</script>

<template>
  <article class="carte" :class="{ actuelle: actuel, epinglee: entree.pinned }">
    <button class="corps" type="button" title="Cliquer pour copier" @click="emit('copier')">
      <span class="pastille" :data-type="entree.kind"><Icone :nom="entree.kind" :taille="15" /></span>
      <span class="contenu">
        <span class="texte">{{ apercu }}</span>
        <span class="meta">
          <span v-if="actuel" class="pret"><Icone nom="coche" :taille="12" /> Prêt à coller · {{ raccourci }}</span>
          <span v-else>{{ libelle }}<template v-if="moment"> · {{ moment }}</template></span>
        </span>
      </span>
    </button>
    <div class="actions">
      <button v-if="ouvrable" class="action" type="button" :title="entree.kind === 'email' ? 'Écrire un courriel' : 'Ouvrir le lien'" @click="emit('ouvrir')">
        <Icone nom="ouvrir" :taille="16" />
      </button>
      <button class="action" type="button" title="QR code" @click="emit('qr')">
        <Icone nom="qr" :taille="16" />
      </button>
      <button class="action" :class="{ active: entree.pinned }" type="button" :title="entree.pinned ? 'Désépingler' : 'Épingler'" @click="emit('epingler')">
        <Icone nom="epingle" :taille="16" :plein="entree.pinned" />
      </button>
    </div>
    <Transition name="flash">
      <div v-if="copie" class="flash"><Icone nom="coche" :taille="16" /> Copié</div>
    </Transition>
  </article>
</template>

<style scoped>
.carte {
  position: relative;
  display: flex;
  align-items: stretch;
  border-radius: var(--rayon);
  background: var(--carte);
  border: 1px solid var(--bordure);
  transition: border-color 0.15s, background 0.15s, box-shadow 0.2s, transform 0.1s;
  overflow: hidden;
}
.carte:hover {
  background: var(--carte-survol);
  border-color: var(--bordure-forte);
}
.carte:active {
  transform: scale(0.99);
}
.carte.actuelle {
  border-color: transparent;
  box-shadow: 0 0 0 1.5px var(--accent), 0 6px 20px -8px var(--accent-ombre);
}
.corps {
  flex: 1;
  min-width: 0;
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 10px 12px 10px 10px;
  text-align: left;
  color: inherit;
}
.pastille {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  color: var(--pastille-texte);
  background: var(--pastille-fond);
}
.pastille[data-type='link'] {
  color: var(--type-lien);
  background: color-mix(in srgb, var(--type-lien) 15%, transparent);
}
.pastille[data-type='email'] {
  color: var(--type-courriel);
  background: color-mix(in srgb, var(--type-courriel) 15%, transparent);
}
.pastille[data-type='phone'] {
  color: var(--type-tel);
  background: color-mix(in srgb, var(--type-tel) 15%, transparent);
}
.contenu {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.texte {
  font-size: 13.5px;
  line-height: 1.38;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
  line-clamp: 3;
  overflow: hidden;
}
.meta {
  font-size: 11.5px;
  color: var(--discret);
}
.pret {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--accent);
  font-weight: 600;
}
/* Les actions n'apparaissent qu'au survol : la liste reste compacte et lisible. */
.actions {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 0 8px 0 32px;
  background: linear-gradient(90deg, transparent, var(--carte-survol-plein) 30px);
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.15s;
}
.carte:hover .actions,
.actions:focus-within {
  opacity: 1;
  pointer-events: auto;
}
.action {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border-radius: 7px;
  color: var(--discret);
}
.action:hover {
  color: var(--texte);
  background: var(--survol);
}
.action.active {
  color: var(--accent);
}
.flash {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, var(--accent-1), var(--accent-2));
  pointer-events: none;
}
.flash-enter-active {
  transition: opacity 0.12s ease-out;
}
.flash-leave-active {
  transition: opacity 0.35s ease-in;
}
.flash-enter-from,
.flash-leave-to {
  opacity: 0;
}
</style>
