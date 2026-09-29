# ClipQR

Historique du presse-papiers pour Windows, macOS et Linux : les 100 derniers textes, fichiers et images copiés, un clic pour les recopier, un QR code pour emporter un texte sur le téléphone, et Impr. écran qui enregistre les captures à la chaîne. Application native en Go ([Wails 3](https://v3.wails.io)), interface en Vue.

## Utilisation

Lancer ClipQR : son icône apparaît près de l’horloge (sous Windows, éventuellement dans les icônes masquées ; sous macOS, dans la barre des menus). Un clic sur l’icône ouvre le panneau à côté de l’horloge.

- **Le panneau reste ouvert** et au premier plan tant qu’on ne le réduit pas : on peut enchaîner clic sur un élément, puis Ctrl+V (⌘V) dans une autre application, puis l’élément suivant.
- **Cliquer sur un élément le recopie.** Il garde sa place dans la liste et porte la marque « Prêt à coller ».
- **Fichiers** : copiés dans l’explorateur (Explorateur Windows, Finder, Nemo, Nautilus, Dolphin…), ils reviennent avec leur nom et leur dossier ; recollés dans un dossier, ils y sont **copiés**, jamais déplacés. ClipQR garde les chemins, pas les fichiers : un fichier déplacé ou supprimé depuis est signalé « Introuvable ».
- **Images** (capture, image d’une page web…) : gardées avec leur vignette, recollées telles quelles. Quand une application offre à la fois du texte et une image (cellules d’un tableur), c’est le texte qui est gardé.
- **Au survol d’un élément** : afficher le **QR code** d’un texte, 🔑 le **masquer**, 📌 l’**épingler**, ou **ouvrir** un lien (navigateur) ou un courriel (messagerie).
- **Mots de passe masqués** : la clé remplace le texte par des points, toujours de la même longueur, pour qui regarde par-dessus l’épaule ; un clic le copie en clair. Elle cache aussi les noms de fichiers et les vignettes. Ce qui est masqué n’est pas transmis à l’interface ; il reste en clair dans le fichier d’historique.
- **Impr. écran par ClipQR** : le bouton 📷 à côté de la pause (ou le menu de l’icône) donne la touche Impr. écran à ClipQR. Chaque appui enregistre tout l’écran dans `Images/ecrans` sous le nom `IE_XXXX_26_09_28_13h32.png` (quatre caractères au hasard, puis la date et l’heure), le met dans le presse-papiers et l’ajoute à la liste : les captures s’enchaînent sans rien ouvrir. Le panneau s’efface le temps de la capture. Un nouveau clic sur 📷 rend la touche au système ; sous Linux, ClipQR libère le raccourci que Cinnamon ou GNOME réserve à Impr. écran et le lui rend en sortant du mode, en quittant, ou au lancement suivant après un arrêt brutal. Non disponible sur macOS (pas de touche Impr. écran) ni sous Wayland.
- **Épinglés** : en tête de liste, hors de la limite des 100, jamais effacés par « Effacer l’historique ».
- **Types détectés** : lien, courriel, téléphone ou texte, avec l’heure de la copie (« il y a 3 min », « hier à 14 h 05 »).
- Rechercher (sans tenir compte des accents), suspendre la collecte, effacer l’historique depuis le panneau.
- **Réduire** : le bouton — du panneau, ou un nouveau clic sur l’icône. Pour arrêter ClipQR : clic droit sur l’icône → **Quitter**.

Une seule instance par session : relancer ClipQR ouvre le panneau de celle qui tourne.

## Données

L’historique est conservé en texte clair sur le poste, sans aucun transfert réseau :

| Système | Fichier |
|---|---|
| Windows | `%APPDATA%\ClipQR\history.json` |
| macOS | `~/Library/Application Support/ClipQR/history.json` |
| Linux | `~/.config/ClipQR/history.json` |

L’historique de la première version Windows est repris tel quel. Les textes copiés pendant la collecte sont enregistrés, y compris les informations confidentielles ; la suspension ignore les copies faites pendant la pause. Les images copiées sont gardées dans le dossier `images` à côté de l’historique et supprimées quand elles en sortent ; les captures Impr. écran restent dans `Images/ecrans`, même après « Effacer l’historique ». Options de lancement : `-data-dir <dossier>`, `-captures-dir <dossier>` et `-paused`.

Limites : textes vides ou de plus de 256 Kio ignorés, images de plus de 48 Mo aussi ; un texte trop long pour un QR code l’indique ; « Effacer l’historique » ne vide pas le presse-papiers du système. Sous Linux, la collecte fonctionne sous X11 et sous Wayland quand le compositeur offre *data-control* (KDE Plasma, GNOME 49 et plus, Sway, Hyprland…) ; sous GNOME, l’icône demande l’extension AppIndicator (présente par défaut sous Ubuntu).

## Installer

Les exécutables sont joints aux [releases](https://github.com/demospec-projects/clipqr/releases), avec leurs empreintes dans `SHA256SUMS.txt`. Ils ne sont pas signés.

- **Windows 10/11** : `ClipQR-windows-amd64.exe`, autonome (le moteur WebView2 est fourni par Windows).
- **macOS** : `ClipQR-macos-universal.zip` (Intel et Apple Silicon). Au premier lancement, clic droit sur ClipQR → **Ouvrir**, puisque l’application n’est pas notariée.
- **Linux** : `ClipQR-linux-amd64.deb` (installe GTK 4 et WebKitGTK 6 au besoin), ou l’exécutable seul dans `ClipQR-linux-amd64.tar.gz` (demande `libgtk-4-1` et `libwebkitgtk-6.0-4`).

## Développer

Prérequis : Go 1.26 ou plus, Node 22 et pnpm, l’outil Wails :

```sh
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.26
```

Sous Linux, les bibliothèques de développement : `sudo apt install libgtk-4-dev libwebkitgtk-6.0-dev`.

```sh
wails3 dev          # l’application avec rechargement à chaud
wails3 build        # bin/clipqr (bin/clipqr.exe sous Windows)
wails3 task test    # tests Go : historique, types, images, noms des captures
```

Windows se compile aussi depuis Linux ou macOS : `wails3 build GOOS=windows`. macOS et Linux se compilent chacun sur leur système ; la compilation GitHub Actions (`.github/workflows/compilation.yml`) produit les trois à chaque étiquette `v*` et les publie en release.

Structure : `main.go` (fenêtre, icône près de l’horloge, menu), `clipservice.go` (lecture du presse-papiers, méthodes appelées par l’interface), `capture.go` (mode Impr. écran), `changes_*`, `printkey_*`, `grab_screen*`, `files_*` (ce qui diffère d’un système à l’autre), `internal/history` (historique, épingles, types, lecture de l’ancien format), `internal/imagestore` (images copiées et vignettes), `internal/screenshot` (nom des captures), `frontend/` (panneau Vue), `build/` (métadonnées et paquets par système). Les icônes se redessinent avec `python3 build/icone.py`, puis `wails3 task common:generate:icons`.
