Première version de ClipQR pour Windows x64.

### Télécharger et utiliser

Téléchargez **ClipQR.exe** dans les fichiers joints à cette release, puis lancez-le. Aucune installation de Go n’est nécessaire.

- Historique local des 100 derniers textes copiés, avec recherche.
- Un clic sur un texte le recopie et masque la fenêtre ; utilisez ensuite Ctrl+V.
- Le bouton **QR code** affiche le texte à scanner avec le téléphone, sans connexion réseau.
- Retrouvez ClipQR dans les icônes près de l’horloge. Clic droit → **Quitter** pour arrêter l’application.
- Possibilité de suspendre la collecte et d’effacer l’historique.

L’historique est conservé en texte clair dans `%APPDATA%\ClipQR\history.json`. Les images et fichiers ne sont pas enregistrés. Les textes trop longs ne peuvent pas être représentés par un QR code.

L’exécutable n’est pas signé. Le fichier **SHA256SUMS.txt** permet de vérifier son empreinte SHA-256.

Validation : compilation Windows x64 et tests automatisés réussis ; collecte, copie en un clic et affichage QR vérifiés sur Windows pendant le développement.
