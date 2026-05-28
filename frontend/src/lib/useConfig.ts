import { useEffect, useState } from 'react'
import { getConfig } from '../api/client'
import type { Config } from '../api/types'

// Cache em nível de módulo: /config não muda durante a sessão, então busca uma
// vez só e compartilha entre as páginas.
let cached: Config | null = null
let inflight: Promise<Config> | null = null

function loadConfig(): Promise<Config> {
  if (cached) return Promise.resolve(cached)
  if (!inflight) {
    inflight = getConfig().then((c) => {
      cached = c
      inflight = null
      return c
    })
  }
  return inflight
}

// useConfig devolve a config da POC (ou null enquanto carrega).
export function useConfig(): Config | null {
  const [config, setConfig] = useState<Config | null>(cached)

  useEffect(() => {
    if (config) return
    let active = true
    loadConfig()
      .then((c) => active && setConfig(c))
      .catch(() => active && setConfig(null))
    return () => {
      active = false
    }
  }, [config])

  return config
}
