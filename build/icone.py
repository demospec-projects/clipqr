"""Dessine les icônes de ClipQR (Python 3 + Pillow + NumPy).

    python3 build/icone.py

Produit build/appicon.png (application, 1024 px), frontend/public/icone.png
(en-tête du panneau), icons/tray.png
(barre des tâches Windows / Linux) et icons/tray-template.png (barre des
menus macOS : silhouette noire que le système recolore). On dessine en
grand puis on réduit, pour des bords lisses à toutes les tailles.
"""
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw

RACINE = Path(__file__).resolve().parent.parent
BLEU = (37, 99, 235)
VIOLET = (124, 58, 237)
ENCRE = (30, 41, 99)


def degrade(taille):
    """Dégradé diagonal du bleu au violet."""
    axe = np.arange(taille)
    t = (axe[None, :] + axe[:, None]) / (2 * (taille - 1))
    rgb = np.array(BLEU) + (np.array(VIOLET) - np.array(BLEU)) * t[..., None]
    alpha = np.full((taille, taille, 1), 255)
    return Image.fromarray(np.concatenate([rgb, alpha], axis=2).round().astype(np.uint8), "RGBA")


def repere(dessin, x, y, cote, couleur, fond):
    """Un des trois carrés de repérage d'un QR code."""
    dessin.rounded_rectangle((x, y, x + cote, y + cote), radius=cote * 0.22, fill=couleur)
    marge = cote * 0.18
    dessin.rounded_rectangle((x + marge, y + marge, x + cote - marge, y + cote - marge), radius=cote * 0.12, fill=fond)
    marge = cote * 0.34
    dessin.rounded_rectangle((x + marge, y + marge, x + cote - marge, y + cote - marge), radius=cote * 0.08, fill=couleur)


def presse_papiers(dessin, s, papier, encre, fond_repere, fente, avec_modules):
    """Le presse-papiers et son motif QR, dans un carré de côté s."""
    dessin.rounded_rectangle((0.24 * s, 0.20 * s, 0.76 * s, 0.84 * s), radius=0.07 * s, fill=papier)
    dessin.rounded_rectangle((0.38 * s, 0.13 * s, 0.62 * s, 0.25 * s), radius=0.04 * s, fill=papier)
    dessin.rounded_rectangle((0.42 * s, 0.165 * s, 0.58 * s, 0.215 * s), radius=0.02 * s, fill=fente)
    cote = 0.16 * s
    for x, y in ((0.31, 0.33), (0.53, 0.33), (0.31, 0.58)):
        repere(dessin, x * s, y * s, cote, encre, fond_repere)
    if avec_modules:
        module = 0.055 * s
        for x, y in ((0.55, 0.60), (0.63, 0.60), (0.55, 0.68), (0.63, 0.68)):
            dessin.rounded_rectangle((x * s, y * s, x * s + module, y * s + module), radius=module * 0.25, fill=encre)


def icone_couleur(taille, avec_modules=True):
    grand = taille * 4
    image = Image.new("RGBA", (grand, grand), (0, 0, 0, 0))
    masque = Image.new("L", (grand, grand), 0)
    ImageDraw.Draw(masque).rounded_rectangle((0, 0, grand - 1, grand - 1), radius=grand * 0.23, fill=255)
    image.paste(degrade(grand), (0, 0), masque)
    presse_papiers(ImageDraw.Draw(image), grand, (255, 255, 255, 255), ENCRE + (255,), (255, 255, 255, 255), (199, 210, 254, 255), avec_modules)
    return image.resize((taille, taille), Image.LANCZOS)


def icone_gabarit(taille):
    """Silhouette pour macOS : le papier plein, le motif évidé."""
    grand = taille * 4
    image = Image.new("RGBA", (grand, grand), (0, 0, 0, 0))
    presse_papiers(ImageDraw.Draw(image), grand, (0, 0, 0, 255), (0, 0, 0, 255), (0, 0, 0, 0), (0, 0, 0, 0), False)
    return image.resize((taille, taille), Image.LANCZOS)


if __name__ == "__main__":
    (RACINE / "icons").mkdir(exist_ok=True)
    icone_couleur(1024).save(RACINE / "build" / "appicon.png")
    (RACINE / "frontend" / "public").mkdir(exist_ok=True)
    icone_couleur(128).save(RACINE / "frontend" / "public" / "icone.png")
    icone_couleur(64, avec_modules=False).save(RACINE / "icons" / "tray.png")
    icone_gabarit(44).save(RACINE / "icons" / "tray-template.png")
    print("Icônes écrites : build/appicon.png, frontend/public/icone.png, icons/tray.png, icons/tray-template.png")
