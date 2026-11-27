/**
 * Som curto de notificação gerado com Web Audio, sem arquivo de áudio ou rede.
 * O contexto é criado por evento e fechado quando o último tom termina.
 * Se o navegador bloquear o áudio até uma interação, a tentativa é silenciosa
 * e o restante da notificação continua funcionando.
 */

const TONES_HZ = [880, 1174.66] as const
const TONE_SECONDS = 0.12
const GAP_SECONDS = 0.04
const PEAK_VOLUME = 0.08

export function playNotificationSound(): boolean {
  const AudioContextConstructor = window.AudioContext
  if (typeof AudioContextConstructor !== 'function') return false

  try {
    const context = new AudioContextConstructor()
    if (context.state === 'suspended') {
      void context.resume().catch(() => undefined)
    }

    const firstToneAt = context.currentTime + 0.01
    const nodes: Array<OscillatorNode | GainNode> = []

    TONES_HZ.forEach((frequency, index) => {
      const startAt = firstToneAt + index * (TONE_SECONDS + GAP_SECONDS)
      const oscillator = context.createOscillator()
      const gain = context.createGain()

      oscillator.type = 'sine'
      oscillator.frequency.setValueAtTime(frequency, startAt)
      gain.gain.setValueAtTime(0, startAt)
      gain.gain.linearRampToValueAtTime(PEAK_VOLUME, startAt + 0.01)
      gain.gain.linearRampToValueAtTime(0, startAt + TONE_SECONDS)

      oscillator.connect(gain)
      gain.connect(context.destination)
      oscillator.start(startAt)
      oscillator.stop(startAt + TONE_SECONDS)
      nodes.push(oscillator, gain)

      if (index === TONES_HZ.length - 1) {
        oscillator.onended = () => {
          for (const node of nodes) node.disconnect()
          void context.close().catch(() => undefined)
        }
      }
    })

    return true
  } catch {
    return false
  }
}
