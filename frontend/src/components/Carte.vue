<script setup lang="ts">
import { computed } from 'vue'
import type { EntryView } from '../../bindings/github.com/demospec-projects/clipqr/models'
import { MASQUE } from '../masque'
import { quand } from '../temps'
import Icone from './Icone.vue'

const props = defineProps<{
  entree: EntryView
  // L'élément est dans le presse-papiers : un Ctrl+V ailleurs le colle.
  actuel: boolean
  // Il vient d'être copié depuis ClipQR (confirmation brève).
  copie: boolean
  maintenant: number
  raccourci: string
}>()
const emit = defineEmits<{ copier: []; epingler: []; masquer: []; qr: []; ouvrir: [] }>()

const libelles: Record<string, string> = { link: 'Lien', email: 'Courriel', phone: 'Téléphone', text: 'Texte' }
const nombre = new Intl.NumberFormat('fr-CA')

function nomDe(chemin: string): string {
  return chemin.split(/[\\/]/).filter(Boolean).pop() ?? chemin
}
function pluriel(n: number, mot: string): string {
  return `${n} ${mot}${n > 1 ? 's' : ''}`
}

const e = computed(() => props.entree)
const fichiers = computed(() => e.value.files ?? [])
const estTexte = computed(() => e.value.kind !== 'files' && e.value.kind !== 'image')
const ouvrable = computed(() => !e.value.secret && (e.value.kind === 'link' || e.value.kind === 'email'))
const vignette = computed(() => (e.value.kind === 'image' && !e.value.secret ? e.value.thumb : ''))

const icone = computed(() => {
  if (e.value.secret) return 'cle'
  if (e.value.kind === 'files') return e.value.folder ? 'dossier' : fichiers.value.length > 1 ? 'fichiers' : 'fichier'
  return e.value.kind
})

const titre = computed(() => {
  if (e.value.secret) return MASQUE
  if (e.value.kind === 'files') {
    const autres = fichiers.value.length - 1
    return nomDe(fichiers.value[0]) + (autres > 0 ? ` + ${pluriel(autres, 'autre')}` : '')
  }
  if (e.value.kind === 'image') {
    return e.value.capture ? e.value.name : `${nombre.format(e.value.width)} × ${nombre.format(e.value.height)} px`
  }
  return e.value.text.trim() + (e.value.truncated ? '…' : '')
})

const introuvable = computed(() => {
  if (e.value.kind === 'files') return e.value.missing === fichiers.value.length || (e.value.secret && e.value.missing > 0)
  return e.value.kind === 'image' && e.value.missing > 0
})

const libelle = computed(() => {
  const n = fichiers.value.length
  switch (e.value.kind) {
    case 'files':
      if (e.value.secret) return 'Fichiers masqués'
      if (introuvable.value) return 'Introuvable'
      if (e.value.missing > 0) return `${pluriel(n, 'fichier')} · ${e.value.missing} introuvable${e.value.missing > 1 ? 's' : ''}`
      return e.value.folder ? 'Dossier' : n > 1 ? pluriel(n, 'fichier') : 'Fichier'
    case 'image':
      if (introuvable.value) return 'Image introuvable'
      if (e.value.secret) return 'Image masquée'
      return e.value.capture ? 'Capture d’écran' : 'Image'
  }
  if (e.value.secret) return 'Mot de passe'
  const lignes = e.value.text.trim().split('\n').length
  return e.value.kind === 'text' && lignes > 1 ? `${lignes} lignes` : libelles[e.value.kind] ?? 'Texte'
})

// Au survol : la liste complète des fichiers, sinon ce que fait le clic.
const infobulle = computed(() => {
  if (e.value.kind === 'files' && !e.value.secret) return fichiers.value.join('\n')
  return e.value.kind === 'image' ? 'Cliquer pour copier l’image' : 'Cliquer pour copier'
})
const moment = computed(() => quand(e.value.copiedAt, props.maintenant))
</script>

<template>
  <article class="carte" :class="{ actuelle: actuel, epinglee: entree.pinned, secrete: entree.secret }">
    <button class="corps" type="button" :title="infobulle" @click="emit('copier')">
      <img v-if="vignette" class="vignette" :src="vignette" alt="" />
      <span v-else class="pastille" :data-type="entree.secret ? 'secret' : entree.kind"><Icone :nom="icone" :taille="15" /></span>
      <span class="contenu">
        <span class="texte" :class="{ ligne: !estTexte }">{{ titre }}</span>
        <span v-if="entree.place" class="lieu">{{ entree.place }}</span>
        <span class="meta">
          <span v-if="actuel" class="pret"><Icone nom="coche" :taille="12" /> Prêt à coller · {{ raccourci }}</span>
          <span v-else :class="{ introuvable }">{{ libelle }}<template v-if="moment"> · {{ moment }}</template></span>
        </span>
      </span>
    </button>
    <div class="actions">
      <button v-if="ouvrable" class="action" type="button" :title="entree.kind === 'email' ? 'Écrire un courriel' : 'Ouvrir le lien'" @click="emit('ouvrir')">
        <Icone nom="ouvrir" :taille="16" />
      </button>
      <button v-if="estTexte" class="action" type="button" title="QR code" @click="emit('qr')">
        <Icone nom="qr" :taille="16" />
      </button>
      <button class="action" :class="{ active: entree.secret }" type="button" :title="entree.secret ? 'Afficher' : 'Masquer (mot de passe)'" @click="emit('masquer')">
        <Icone nom="cle" :taille="16" />
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
.pastille[data-type='files'] {
  color: var(--type-fichier);
  background: color-mix(in srgb, var(--type-fichier) 15%, transparent);
}
.pastille[data-type='image'] {
  color: var(--type-image);
  background: color-mix(in srgb, var(--type-image) 15%, transparent);
}
.vignette {
  flex: none;
  width: 64px;
  height: 44px;
  object-fit: cover;
  border-radius: 8px;
  border: 1px solid var(--bordure);
  background: var(--pastille-fond);
}
.texte.ligne {
  -webkit-line-clamp: 1;
  line-clamp: 1;
  word-break: break-all;
}
.lieu {
  font-size: 11.5px;
  color: var(--discret);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.introuvable {
  color: var(--attention);
  font-weight: 600;
}
.pastille[data-type='secret'] {
  color: var(--type-secret);
  background: color-mix(in srgb, var(--type-secret) 15%, transparent);
}
.secrete .texte {
  font-size: 16px;
  line-height: 1.2;
  letter-spacing: 0.14em;
  color: var(--discret);
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
