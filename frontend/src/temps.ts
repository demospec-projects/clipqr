const relatif = new Intl.RelativeTimeFormat('fr-CA', { numeric: 'auto', style: 'short' })
const heure = new Intl.DateTimeFormat('fr-CA', { hour: 'numeric', minute: '2-digit' })
const jour = new Intl.DateTimeFormat('fr-CA', { day: 'numeric', month: 'short' })

function debutDuJour(t: number): number {
  const d = new Date(t)
  return new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime()
}

// « à l’instant », « il y a 3 min », « il y a 2 h », « hier à 14 h 05 », « 12 sept. »
export function quand(iso: string | null, maintenant: number): string {
  if (!iso) return ''
  const t = new Date(iso).getTime()
  const secondes = Math.max(0, Math.round((maintenant - t) / 1000))
  if (secondes < 45) return 'à l’instant'
  const minutes = Math.round(secondes / 60)
  if (minutes < 60) return relatif.format(-minutes, 'minute')
  const heures = Math.round(minutes / 60)
  if (heures < 12 || debutDuJour(t) === debutDuJour(maintenant)) return relatif.format(-heures, 'hour')
  const hier = new Date(maintenant)
  hier.setDate(hier.getDate() - 1)
  if (debutDuJour(t) === debutDuJour(hier.getTime())) return `hier à ${heure.format(t)}`
  return jour.format(t)
}
