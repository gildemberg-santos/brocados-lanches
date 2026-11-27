import { afterEach, describe, expect, it, vi } from 'vitest'

import { playNotificationSound } from './notificationSound'

class FakeAudioParam {
  setValueAtTime = vi.fn()
  linearRampToValueAtTime = vi.fn()
}

class FakeAudioNode {
  connect = vi.fn()
  disconnect = vi.fn()
}

class FakeOscillator extends FakeAudioNode {
  type = ''
  frequency = new FakeAudioParam()
  start = vi.fn()
  stop = vi.fn()
  onended: (() => void) | null = null
}

class FakeGain extends FakeAudioNode {
  gain = new FakeAudioParam()
}

class FakeAudioContext {
  state: AudioContextState
  currentTime = 0
  destination = new FakeAudioNode()
  oscillators: FakeOscillator[] = []
  gains: FakeGain[] = []
  resume = vi.fn().mockResolvedValue(undefined)
  close = vi.fn().mockResolvedValue(undefined)

  constructor(state: AudioContextState = 'running') {
    this.state = state
    FakeAudioContext.instances.push(this)
  }

  static instances: FakeAudioContext[] = []

  createOscillator(): FakeOscillator {
    const oscillator = new FakeOscillator()
    this.oscillators.push(oscillator)
    return oscillator
  }

  createGain(): FakeGain {
    const gain = new FakeGain()
    this.gains.push(gain)
    return gain
  }
}

const original = Object.getOwnPropertyDescriptor(window, 'AudioContext')

function installAudioContext(context: FakeAudioContext): void {
  Object.defineProperty(window, 'AudioContext', {
    configurable: true,
    value: class {
      constructor() {
        return context
      }
    },
  })
}

describe('playNotificationSound', () => {
  afterEach(() => {
    FakeAudioContext.instances = []
    if (original) {
      Object.defineProperty(window, 'AudioContext', original)
    } else {
      Reflect.deleteProperty(window, 'AudioContext')
    }
  })

  it('não falha quando o navegador não oferece Web Audio', () => {
    Reflect.deleteProperty(window, 'AudioContext')

    expect(playNotificationSound()).toBe(false)
  })

  it('toca dois tons curtos ligados à saída de áudio', () => {
    const context = new FakeAudioContext()
    installAudioContext(context)

    expect(playNotificationSound()).toBe(true)
    expect(context.oscillators).toHaveLength(2)
    expect(context.oscillators[0]?.frequency.setValueAtTime).toHaveBeenCalledWith(880, 0.01)
    expect(context.oscillators[1]?.frequency.setValueAtTime).toHaveBeenCalledWith(1174.66, 0.17)
    expect(context.oscillators[0]?.start).toHaveBeenCalledWith(0.01)
    expect(context.oscillators[0]?.stop).toHaveBeenCalledWith(0.13)
    expect(context.oscillators[0]?.connect).toHaveBeenCalledWith(context.gains[0])
    expect(context.gains[0]?.connect).toHaveBeenCalledWith(context.destination)
  })

  it('tenta retomar contexto suspenso e libera os nós ao terminar', () => {
    const context = new FakeAudioContext('suspended')
    installAudioContext(context)

    expect(playNotificationSound()).toBe(true)
    expect(context.resume).toHaveBeenCalledTimes(1)

    const lastOscillator = context.oscillators.at(-1)
    lastOscillator?.onended?.()

    expect(lastOscillator?.disconnect).toHaveBeenCalledTimes(1)
    expect(context.gains[1]?.disconnect).toHaveBeenCalledTimes(1)
    expect(context.close).toHaveBeenCalledTimes(1)
  })

  it('não propaga erro quando o contexto não pode ser criado', () => {
    Object.defineProperty(window, 'AudioContext', {
      configurable: true,
      value: class {
        constructor() {
          throw new Error('autoplay bloqueado')
        }
      },
    })

    expect(playNotificationSound()).toBe(false)
  })
})
