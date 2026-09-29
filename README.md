# ClipQR pour Windows

Application native en Go : historique des 100 derniers textes copiés et QR codes utilisables hors ligne.

## Utilisation

Lancer `dist/ClipQR.exe`. Copier du texte dans n’importe quelle application, puis ouvrir ClipQR avec son icône près de l’horloge (éventuellement dans les icônes masquées).

- Cliquer sur le texte pour le recopier : la fenêtre se masque, puis utiliser Ctrl+V dans l’application souhaitée.
- Cliquer sur **QR code** pour afficher un QR à scanner avec le téléphone.
- Rechercher, suspendre la collecte ou effacer l’historique depuis la fenêtre.
- Fermer la fenêtre laisse l’application active. Pour l’arrêter : clic droit sur l’icône → **Quitter**.

Historique local dans `%APPDATA%\ClipQR\history.json`, en texte clair. Les textes copiés pendant la collecte sont enregistrés, y compris les informations confidentielles. La suspension ignore les changements pendant la pause. Aucun transfert réseau n’est effectué par l’application.

L’application collecte le texte courant au démarrage puis surveille les changements toutes les 400 ms. Des copies extrêmement rapprochées peuvent être manquées. Les textes vides ou dépassant 256 Kio sont ignorés. Les images et fichiers ne sont pas enregistrés. Les longs textes peuvent dépasser la capacité du QR code ; un message l’indique. Le bouton Effacer ne vide pas le presse-papiers Windows.

Une seule instance par session Windows. Aucun démarrage automatique n’est configuré.

## Compiler

Go 1.26.4 ou ultérieur, Windows x64. Depuis PowerShell :

```powershell
.\build.ps1
```

Le premier lancement du script nécessite Internet pour télécharger les dépendances. Le manifeste Windows est intégré à l’exécutable. Aucun compilateur C n’est requis. L’exécutable est autonome et n’est pas signé.
