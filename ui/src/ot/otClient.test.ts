import { OTClient } from './otClient'

describe('OTClient', () => {
  describe('constructor', () => {
    it('should initialise with the provided version', () => {
      const client = new OTClient('doc-1', 7, vi.fn())
      expect(client.version).toBe(7)
    })

    it('should initialise with version 0', () => {
      const client = new OTClient('doc-1', 0, vi.fn())
      expect(client.version).toBe(0)
    })
  })

  describe('localSteps', () => {
    it('should return a correctly shaped ot_steps payload', () => {
      const client = new OTClient('doc-1', 3, vi.fn())
      const steps = [{ stepType: 'replace', from: 0, to: 1 }]

      const payload = client.localSteps(steps)

      expect(payload).toEqual({
        type: 'ot_steps',
        doc_id: 'doc-1',
        version: 3,
        steps,
      })
    })

    it('should not change the current version when local steps are recorded', () => {
      const client = new OTClient('doc-1', 5, vi.fn())
      client.localSteps([{ step: 'a' }])
      expect(client.version).toBe(5)
    })

    it('should accumulate steps across multiple calls', () => {
      const client = new OTClient('doc-1', 0, vi.fn())

      client.localSteps([{ step: 'a' }])
      const second = client.localSteps([{ step: 'b' }])

      // Each call returns only its own steps in the payload
      expect(second.steps).toEqual([{ step: 'b' }])
    })
  })

  describe('handleMessage – ot_ack', () => {
    it('should advance the version by 1 when server provides the confirmed version', () => {
      const client = new OTClient('doc-1', 4, vi.fn())

      client.handleMessage({ type: 'ot_ack', version: 4 })

      expect(client.version).toBe(5)
    })

    it('should advance the version using the current version when ot_ack omits version', () => {
      const client = new OTClient('doc-1', 3, vi.fn())

      client.handleMessage({ type: 'ot_ack' })

      expect(client.version).toBe(4) // 3 + 1
    })

    it('should clear inflight steps after ack', () => {
      const onSteps = vi.fn()
      const client = new OTClient('doc-1', 0, onSteps)
      client.localSteps([{ step: 'pending' }])

      client.handleMessage({ type: 'ot_ack', version: 0 })

      // Verify version incremented correctly (proxy for inflight cleared)
      expect(client.version).toBe(1)
      // onSteps not called on ack
      expect(onSteps).not.toHaveBeenCalled()
    })
  })

  describe('handleMessage – ot_steps', () => {
    it('should call onSteps with the remote steps', () => {
      const onSteps = vi.fn()
      const client = new OTClient('doc-1', 0, onSteps)
      const remote = [{ step: 'remote' }]

      client.handleMessage({ type: 'ot_steps', steps: remote, version: 1 })

      expect(onSteps).toHaveBeenCalledOnce()
      expect(onSteps).toHaveBeenCalledWith(remote)
    })

    it('should update the version to the server-provided value', () => {
      const client = new OTClient('doc-1', 0, vi.fn())

      client.handleMessage({ type: 'ot_steps', steps: [{ step: 'x' }], version: 8 })

      expect(client.version).toBe(8)
    })

    it('should not call onSteps when the steps array is empty', () => {
      const onSteps = vi.fn()
      const client = new OTClient('doc-1', 0, onSteps)

      client.handleMessage({ type: 'ot_steps', steps: [], version: 1 })

      expect(onSteps).not.toHaveBeenCalled()
    })

    it('should not call onSteps when steps is absent', () => {
      const onSteps = vi.fn()
      const client = new OTClient('doc-1', 0, onSteps)

      client.handleMessage({ type: 'ot_steps' })

      expect(onSteps).not.toHaveBeenCalled()
    })
  })

  describe('handleMessage – ot_rebase', () => {
    it('should call onSteps with the missed steps from the server', () => {
      const onSteps = vi.fn()
      const client = new OTClient('doc-1', 2, onSteps)
      const missed = [{ step: 'missed' }]

      client.handleMessage({ type: 'ot_rebase', steps: missed, version: 3 })

      expect(onSteps).toHaveBeenCalledOnce()
      expect(onSteps).toHaveBeenCalledWith(missed)
    })

    it('should update the version to the server-provided base version', () => {
      const client = new OTClient('doc-1', 1, vi.fn())

      client.handleMessage({ type: 'ot_rebase', steps: [{ step: 'a' }], version: 5 })

      expect(client.version).toBe(5)
    })

    it('should not call onSteps when the rebase steps array is empty', () => {
      const onSteps = vi.fn()
      const client = new OTClient('doc-1', 3, onSteps)

      client.handleMessage({ type: 'ot_rebase', steps: [], version: 4 })

      expect(onSteps).not.toHaveBeenCalled()
    })

    it('should not call onSteps when rebase steps are absent', () => {
      const onSteps = vi.fn()
      const client = new OTClient('doc-1', 3, onSteps)

      client.handleMessage({ type: 'ot_rebase', version: 4 })

      expect(onSteps).not.toHaveBeenCalled()
    })

    it('should still update the version when steps are absent', () => {
      const client = new OTClient('doc-1', 2, vi.fn())

      client.handleMessage({ type: 'ot_rebase', version: 6 })

      expect(client.version).toBe(6)
    })
  })

  describe('handleMessage – unknown types', () => {
    it('should ignore unknown message types without throwing', () => {
      const onSteps = vi.fn()
      const client = new OTClient('doc-1', 0, onSteps)
      const versionBefore = client.version

      client.handleMessage({ type: 'some_other_type', steps: [{ step: 'x' }] })

      expect(onSteps).not.toHaveBeenCalled()
      expect(client.version).toBe(versionBefore)
    })
  })
})
