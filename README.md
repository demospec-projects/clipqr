# ClipQR

Historique du presse-papiers pour Windows, macOS et Linux : les 100 derniers textes copiés, un clic pour les recopier, un QR code pour les emporter sur le téléphone. Application native en Go ([Wails 3](https://v3.wails.io)), interface en Vue.

## Utilisation

Lancer ClipQR : son icône apparaît près de l’horloge (sous Windows, éventuellement dans les icônes masquées ; sous macOS, dans la barre des menus). Un clic sur l’icône ouvre le panneau à côté de l’horloge.

- **Le panneau reste ouvert** et au premier plan tant qu’on ne le réduit pas : on peut enchaîner clic sur un texte, puis Ctrl+V (⌘V) dans une autre application, puis le texte suivant.
- **Cliquer sur un texte le copie.** Il garde sa place dans la liste et porte la marque « Prêt à coller ».
- **Au survol d’un texte** : 📌 l’épingler, afficher son **QR code**, ou **ouvrir** un lien (navigateur) ou un courriel (messagerie).
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

L’historique de la première version Windows est repris tel quel. Les textes copiés pendant la collecte sont enregistrés, y compris les informations confidentielles ; la suspension ignore les copies faites pendant la pause. Options de lancement : `-data-dir <dossier>` et `-paused`.

Limites : textes vides ou de plus de 256 Kio ignorés ; images et fichiers non enregistrés ; un texte trop long pour un QR code l’indique ; « Effacer l’historique » ne vide pas le presse-papiers du système. Sous Linux, la collecte fonctionne sous X11 et sous Wayland quand le compositeur offre *data-control* (KDE Plasma, GNOME 49 et plus, Sway, Hyprland…) ; sous GNOME, l’icône demande l’extension AppIndicator (présente par défaut sous Ubuntu).

## Installer

Les exécutables sont joints aux [releases](https://github.com/demospec-projects/clipqr/releases), avec leurs empreintes dans `SHA256SUMS.txt`. Ils ne sont pas signés.

- **Windows 10/11** : `ClipQR-windows-amd64.exe`, autonome (le moteur WebView2 est fourni par Windows).
- **macOS** : `ClipQR-macos-universal.zip` (Intel et Apple Silicon). Au premier lancement, clic droit sur ClipQR → **Ouvrir**, puisque l’application n’est pas notariée.
- **Linux** : `ClipQR-linux-amd64.deb` (installe GTK 4 et WebKitGTK 6 au besoin), ou l’exécutable seul dans `ClipQR-linux-amd64.tar.gz` (demande `libgtk-4-1` et `libwebkitgtk-6.0-4`).

## Développer

Prérequis : Go 1.25 ou plus, Node 22 et pnpm, l’outil Wails :

```sh
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.26
```

Sous Linux, les bibliothèques de développement : `sudo apt install libgtk-4-dev libwebkitgtk-6.0-dev`.

```sh
wails3 dev          # l’application avec rechargement à chaud
wails3 build        # bin/clipqr (bin/clipqr.exe sous Windows)
wails3 task test    # tests Go de l’historique et des types
```

Windows se compile aussi depuis Linux ou macOS : `wails3 build GOOS=windows`. macOS et Linux se compilent chacun sur leur système ; la compilation GitHub Actions (`.github/workflows/compilation.yml`) produit les trois à chaque étiquette `v*` et les publie en release.

Structure : `main.go` (fenêtre, icône près de l’horloge, menu), `clipservice.go` (surveillance du presse-papiers, méthodes appelées par l’interface), `internal/history` (historique, épingles, types, lecture de l’ancien format), `frontend/` (panneau Vue), `build/` (métadonnées et paquets par système). Les icônes se redessinent avec `python3 build/icone.py`, puis `wails3 task common:generate:icons`.
