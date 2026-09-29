<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { Events, Window } from '@wailsio/runtime'
import * as Clip from '../bindings/github.com/demospec-projects/clipqr/clipservice'
import type { EntryView, State } from '../bindings/github.com/demospec-projects/clipqr/models'
import Carte from './components/Carte.vue'
import FeuilleQr from './components/FeuilleQr.vue'
import Icone from './components/Icone.vue'
import { MASQUE } from './masque'

const etat = ref<State>({ entries: [], paused: false, current: '', error: '', captureMode: false, captureAvailable: false, capturesDir: '' })
const recherche = ref('')
const maintenant = ref(Date.now())
const copieRecente = ref('')
const confirmerEffacement = ref(false)
const qr = ref<{ entree: EntryView; image: string; erreur: string } | null>(null)
const erreurAction = ref('')

const raccourci = navigator.userAgent.includes('Mac') ? '⌘V' : 'Ctrl+V'

// La recherche ignore la casse et les accents : « evenement » trouve « Événement ».
function sansAccents(texte: string): string {
  return texte.normalize('NFD').replace(/\p{Diacritic}/gu, '').toLowerCase()
}
const entrees = computed(() => etat.value.entries ?? [])
// On cherche dans le texte, dans les noms et le dossier des fichiers, et dans le nom des captures.
function cleDeRecherche(e: EntryView): string {
  return sansAccents([e.text, ...(e.files ?? []), e.name].join(' '))
}
const cles = computed(() => new Map(entrees.value.map((e) => [e.id, cleDeRecherche(e)])))
const visibles = computed(() => {
  const q = sansAccents(recherche.value.trim())
  return q ? entrees.value.filter((e) => cles.value.get(e.id)?.includes(q)) : entrees.value
})
const epinglees = computed(() => visibles.value.filter((e) => e.pinned))
const recentes = computed(() => visibles.value.filter((e) => !e.pinned))
const nonEpinglees = computed(() => entrees.value.filter((e) => !e.pinned).length)

let arretEcoute: (() => void) | undefined
let horloge: number | undefined
let minuterieCopie: number | undefined

onMounted(async () => {
  arretEcoute = Events.On('clipqr:state', (ev) => {
    etat.value = ev.data
  })
  horloge = window.setInterval(() => (maintenant.value = Date.now()), 30_000)
  etat.value = await Clip.State()
})
onBeforeUnmount(() => {
  arretEcoute?.()
  window.clearInterval(horloge)
})

async function agir(action: () => Promise<unknown>) {
  try {
    erreurAction.value = ''
    await action()
  } catch (e) {
    erreurAction.value = e instanceof Error ? e.message : String(e)
  }
}

function copier(entree: EntryView) {
  agir(async () => {
    await Clip.Copy(entree.id)
    copieRecente.value = entree.id
    window.clearTimeout(minuterieCopie)
    minuterieCopie = window.setTimeout(() => (copieRecente.value = ''), 650)
  })
}

async function afficherQr(entree: EntryView) {
  qr.value = { entree, image: '', erreur: '' }
  try {
    const image = await Clip.QRCode(entree.id)
    if (qr.value?.entree.id === entree.id) qr.value.image = image
  } catch (e) {
    if (qr.value?.entree.id === entree.id) qr.value.erreur = e instanceof Error ? e.message : String(e)
  }
}

function effacer() {
  confirmerEffacement.value = false
  agir(() => Clip.Clear())
}

function surToucheRecherche(e: KeyboardEvent) {
  if (e.key === 'Escape') recherche.value = ''
}
</script>

<template>
  <main class="panneau">
    <header class="entete">
      <div class="marque">
        <img src="/icone.png" alt="" width="30" height="30" />
        <div>
          <h1>ClipQR</h1>
          <p>Un clic pour copier. Un scan pour emporter.</p>
        </div>
      </div>
      <div class="boutons">
        <button
          class="bouton"
          :class="{ alerte: etat.paused }"
          type="button"
          :title="etat.paused ? 'Reprendre la collecte' : 'Suspendre la collecte'"
          @click="agir(() => Clip.SetPaused(!etat.paused))"
        >
          <Icone :nom="etat.paused ? 'reprendre' : 'pause'" :taille="16" />
        </button>
        <button
          v-if="etat.captureAvailable"
          class="bouton"
          :class="{ actif: etat.captureMode }"
          type="button"
          :title="
            etat.captureMode
              ? `Impr. écran enregistre dans ${etat.capturesDir} — cliquer pour rendre la touche au système`
              : 'Prendre Impr. écran : chaque capture est enregistrée et gardée ici'
          "
          @click="agir(() => Clip.SetCaptureMode(!etat.captureMode))"
        >
          <Icone nom="camera" :taille="16" />
        </button>
        <button class="bouton" type="button" title="Effacer l’historique" :disabled="nonEpinglees === 0" @click="confirmerEffacement = true">
          <Icone nom="corbeille" :taille="16" />
        </button>
        <button class="bouton" type="button" title="Réduire" @click="Window.Hide()">
          <Icone nom="reduire" :taille="16" />
        </button>
      </div>
    </header>

    <div class="recherche">
      <Icone nom="recherche" :taille="16" />
      <input v-model="recherche" type="search" placeholder="Rechercher dans l’historique…" spellcheck="false" @keydown="surToucheRecherche" />
      <button v-if="recherche" class="vider" type="button" title="Vider la recherche" @click="recherche = ''">
        <Icone nom="fermer" :taille="14" />
      </button>
    </div>

    <Transition name="bandeau">
      <div v-if="confirmerEffacement" class="bandeau confirmation">
        <span>Effacer {{ nonEpinglees }} élément{{ nonEpinglees > 1 ? 's' : '' }} ? Les épinglés et les captures enregistrées restent.</span>
        <button type="button" class="lien" @click="confirmerEffacement = false">Annuler</button>
        <button type="button" class="danger" @click="effacer">Effacer</button>
      </div>
    </Transition>
    <div v-if="etat.error || erreurAction" class="bandeau erreur">
      <Icone nom="alerte" :taille="16" />
      <span>{{ erreurAction || etat.error }}</span>
      <button type="button" class="lien" title="Masquer" @click="erreurAction ? (erreurAction = '') : agir(() => Clip.DismissError())">
        <Icone nom="fermer" :taille="14" />
      </button>
    </div>

    <section class="liste">
      <template v-if="epinglees.length">
        <h2>Épinglés <span>{{ epinglees.length }}</span></h2>
        <TransitionGroup name="carte" tag="div" class="cartes">
          <Carte
            v-for="e in epinglees"
            :key="e.id"
            :entree="e"
            :actuel="e.id === etat.current"
            :copie="e.id === copieRecente"
            :maintenant="maintenant"
            :raccourci="raccourci"
            @copier="copier(e)"
            @epingler="agir(() => Clip.SetPinned(e.id, false))"
            @masquer="agir(() => Clip.SetSecret(e.id, !e.secret))"
            @qr="afficherQr(e)"
            @ouvrir="agir(() => Clip.Open(e.id))"
          />
        </TransitionGroup>
      </template>

      <template v-if="recentes.length">
        <h2 v-if="epinglees.length">Récents <span>{{ recentes.length }}</span></h2>
        <TransitionGroup name="carte" tag="div" class="cartes">
          <Carte
            v-for="e in recentes"
            :key="e.id"
            :entree="e"
            :actuel="e.id === etat.current"
            :copie="e.id === copieRecente"
            :maintenant="maintenant"
            :raccourci="raccourci"
            @copier="copier(e)"
            @epingler="agir(() => Clip.SetPinned(e.id, true))"
            @masquer="agir(() => Clip.SetSecret(e.id, !e.secret))"
            @qr="afficherQr(e)"
            @ouvrir="agir(() => Clip.Open(e.id))"
          />
        </TransitionGroup>
      </template>

      <div v-if="!visibles.length" class="vide">
        <Icone :nom="recherche ? 'recherche' : 'pressePapiers'" :taille="34" />
        <p v-if="recherche">Aucun texte ne correspond à « {{ recherche }} ».</p>
        <p v-else>Copiez un texte, des fichiers ou une image pour commencer ({{ raccourci.replace('V', 'C') }}).</p>
      </div>
    </section>

    <footer class="pied">
      <span class="statut" :class="{ suspendu: etat.paused }">
        <i />{{ etat.paused ? 'Collecte suspendue' : 'Collecte active' }}
      </span>
      <span v-if="etat.captureMode" class="captures" :title="`Chaque Impr. écran est enregistrée dans ${etat.capturesDir}`">
        <Icone nom="camera" :taille="13" /> Impr. écran → {{ etat.capturesDir }}
      </span>
      <span v-else>{{ entrees.length }} élément{{ entrees.length > 1 ? 's' : '' }} · conservés sur ce poste</span>
    </footer>

    <FeuilleQr
      v-if="qr"
      :apercu="qr.entree.secret ? MASQUE : qr.entree.text.trim()"
      :image="qr.image"
      :erreur="qr.erreur"
      @fermer="qr = null"
    />
  </main>
</template>

<style scoped>
.panneau {
  position: relative;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--fond);
  overflow: hidden;
}
.entete {
  --wails-draggable: drag;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 12px 10px 16px;
  background: var(--entete);
}
.marque {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}
.marque img {
  border-radius: 8px;
  box-shadow: 0 4px 14px -4px var(--accent-ombre);
}
h1 {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  letter-spacing: -0.01em;
}
.marque p {
  margin: 1px 0 0;
  font-size: 11.5px;
  color: var(--discret);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.boutons {
  --wails-draggable: no-drag;
  display: flex;
  flex: none;
  gap: 1px;
}
.bouton {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border-radius: 9px;
  color: var(--discret);
  transition: background 0.15s, color 0.15s;
}
.bouton:hover:not(:disabled) {
  background: var(--survol);
  color: var(--texte);
}
.bouton:disabled {
  opacity: 0.35;
}
.bouton.actif {
  color: #fff;
  background: linear-gradient(135deg, var(--accent-1), var(--accent-2));
  box-shadow: 0 4px 12px -4px var(--accent-ombre);
}
.bouton.alerte {
  color: var(--attention);
  background: color-mix(in srgb, var(--attention) 14%, transparent);
}
.recherche {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 2px 12px 8px;
  padding: 0 10px;
  height: 38px;
  border-radius: 11px;
  background: var(--champ);
  border: 1px solid var(--bordure);
  color: var(--discret);
  transition: border-color 0.15s, box-shadow 0.15s;
}
.recherche:focus-within {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent) 22%, transparent);
}
.recherche input {
  flex: 1;
  min-width: 0;
  border: 0;
  outline: 0;
  background: none;
  color: var(--texte);
  font: inherit;
  font-size: 13.5px;
}
.recherche input::-webkit-search-cancel-button {
  display: none;
}
.vider {
  display: grid;
  place-items: center;
  width: 22px;
  height: 22px;
  border-radius: 6px;
  color: var(--discret);
}
.vider:hover {
  background: var(--survol);
}
.bandeau {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 12px 8px;
  padding: 8px 10px;
  border-radius: 10px;
  font-size: 12.5px;
}
.bandeau span {
  flex: 1;
}
.confirmation {
  background: color-mix(in srgb, var(--danger) 12%, transparent);
  border: 1px solid color-mix(in srgb, var(--danger) 30%, transparent);
}
.erreur {
  color: var(--attention);
  background: color-mix(in srgb, var(--attention) 12%, transparent);
}
.erreur span {
  color: var(--texte);
}
.lien {
  padding: 4px 8px;
  border-radius: 7px;
  color: var(--discret);
}
.lien:hover {
  background: var(--survol);
  color: var(--texte);
}
.danger {
  padding: 5px 10px;
  border-radius: 7px;
  font-weight: 600;
  color: #fff;
  background: var(--danger);
}
.liste {
  flex: 1;
  overflow-y: auto;
  padding: 2px 12px 12px;
  scrollbar-gutter: stable;
}
h2 {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 8px 2px 6px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--discret);
}
h2 span {
  padding: 0 6px;
  border-radius: 99px;
  background: var(--survol);
  font-size: 10.5px;
  letter-spacing: 0;
}
.cartes {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.vide {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 56px 24px;
  text-align: center;
  color: var(--discret);
}
.vide p {
  margin: 0;
  font-size: 13px;
}
.pied {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 9px 16px;
  border-top: 1px solid var(--bordure);
  font-size: 11.5px;
  color: var(--discret);
}
.captures {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  color: var(--accent);
  font-weight: 600;
}
.statut {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font-weight: 600;
  color: var(--succes);
}
.statut i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 0 0 currentColor;
  animation: pouls 2.4s ease-out infinite;
}
.statut.suspendu {
  color: var(--attention);
}
.statut.suspendu i {
  animation: none;
}
.carte-enter-active,
.carte-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}
.carte-enter-from {
  opacity: 0;
  transform: translateY(-6px);
}
.carte-leave-to {
  opacity: 0;
  transform: translateX(12px);
}
.carte-leave-active {
  position: absolute;
  width: 100%;
}
.carte-move {
  transition: transform 0.25s;
}
.bandeau-enter-active,
.bandeau-leave-active {
  transition: opacity 0.15s, transform 0.15s;
}
.bandeau-enter-from,
.bandeau-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
@keyframes pouls {
  0% {
    box-shadow: 0 0 0 0 color-mix(in srgb, currentColor 55%, transparent);
  }
  70%,
  100% {
    box-shadow: 0 0 0 6px transparent;
  }
}
</style>
